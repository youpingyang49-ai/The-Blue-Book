package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	DefaultAccessTTL  = 15 * time.Minute
	DefaultRefreshTTL = 7 * 24 * time.Hour
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	if accessTTL <= 0 {
		accessTTL = DefaultAccessTTL
	}
	if refreshTTL <= 0 {
		refreshTTL = DefaultRefreshTTL
	}
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

type Claims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone,omitempty"`
	Type   string `json:"typ,omitempty"`
	// FamilyID 标识一条 refresh token 轮换链（令牌家族），
	// 同一家族内只有最新一次签发的 refresh token 有效。
	FamilyID string `json:"fid,omitempty"`
	// Generation 是家族内的轮换代数，仅用于排查问题。
	Generation int `json:"gen,omitempty"`
	jwt.RegisteredClaims
}

// RefreshToken 是一次签发的 refresh token 及其 Redis 侧所需元数据。
type RefreshToken struct {
	Token      string
	JTI        string
	FamilyID   string
	Generation int
	ExpiresAt  time.Time
}

// NewFamilyID 生成新的令牌家族 ID。
func NewFamilyID() (string, error) {
	return randomID()
}

func (m *Manager) IssueAccessToken(userID uint, phone, familyID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Phone:    phone,
		Type:     TokenTypeAccess,
		FamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) IssueRefreshToken(userID uint, familyID string, generation int) (*RefreshToken, error) {
	jti, err := randomID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiresAt := now.Add(m.refreshTTL)
	claims := Claims{
		UserID:     userID,
		Type:       TokenTypeRefresh,
		FamilyID:   familyID,
		Generation: generation,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		Token:      token,
		JTI:        jti,
		FamilyID:   familyID,
		Generation: generation,
		ExpiresAt:  expiresAt,
	}, nil
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ParseAccess 校验 access token（拒绝 refresh token 冒充）。
func (m *Manager) ParseAccess(tokenStr string) (*Claims, error) {
	claims, err := m.Parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ParseRefresh 校验 refresh token（拒绝 access token 冒充）。
func (m *Manager) ParseRefresh(tokenStr string) (*Claims, error) {
	claims, err := m.Parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
