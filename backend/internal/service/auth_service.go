package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/session"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

var (
	ErrCodeExpired    = errors.New("验证码已过期")
	ErrCodeInvalid    = errors.New("验证码错误")
	ErrRateLimited    = errors.New("发送太频繁，请稍后再试")
	ErrSessionExpired = errors.New("登录状态已失效")
	ErrSessionReused  = errors.New("登录凭证已被重复使用")
)

const (
	codeTTL = 5 * time.Minute
	rateTTL = 60 * time.Second
)

type AuthService struct {
	userRepo *repository.UserRepo
	rdb      *redis.Client
	jwtMgr   *jwt.Manager
	sessions *session.Store
}

func NewAuthService(userRepo *repository.UserRepo, rdb *redis.Client, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		rdb:      rdb,
		jwtMgr:   jwtMgr,
		sessions: session.NewStore(rdb, jwtMgr.RefreshTTL()),
	}
}

func (s *AuthService) SendCode(ctx context.Context, phone string) (string, error) {
	rateKey := "sms:rate:" + phone
	exists, err := s.rdb.Exists(ctx, rateKey).Result()
	if err != nil {
		return "", err
	}
	if exists > 0 {
		return "", ErrRateLimited
	}

	code := generateCode()
	codeKey := "sms:code:" + phone
	if err := s.rdb.Set(ctx, codeKey, code, codeTTL).Err(); err != nil {
		return "", err
	}
	if err := s.rdb.Set(ctx, rateKey, "1", rateTTL).Err(); err != nil {
		return "", err
	}
	return code, nil
}

func (s *AuthService) Login(ctx context.Context, phone, code string) (string, string, *model.User, error) {
	codeKey := "sms:code:" + phone
	stored, err := s.rdb.Get(ctx, codeKey).Result()
	if errors.Is(err, redis.Nil) {
		return "", "", nil, ErrCodeExpired
	}
	if err != nil {
		return "", "", nil, err
	}
	if stored != code {
		return "", "", nil, ErrCodeInvalid
	}
	_ = s.rdb.Del(ctx, codeKey)

	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		user = &model.User{
			Phone:    phone,
			Nickname: defaultNickname(phone),
		}
		if err := s.userRepo.Create(user); err != nil {
			return "", "", nil, err
		}
	}

	familyID, err := jwt.NewFamilyID()
	if err != nil {
		return "", "", nil, err
	}
	refresh, err := s.jwtMgr.IssueRefreshToken(user.ID, familyID, 1)
	if err != nil {
		return "", "", nil, err
	}
	access, err := s.jwtMgr.IssueAccessToken(user.ID, user.Phone, familyID)
	if err != nil {
		return "", "", nil, err
	}
	// 先登记家族当前 jti 再返回 token；登记失败说明无法保证单次使用，拒绝发牌。
	if err := s.sessions.Start(ctx, familyID, refresh.JTI); err != nil {
		return "", "", nil, err
	}
	return access, refresh.Token, user, nil
}

// Refresh 用 refresh token 换取新的 access / refresh token。
// 每次轮换都会让旧 refresh token 立即失效（Redis 家族状态），
// 旧 token 再次出现时按「并发刷新」或「凭证重放」处理。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, *model.User, error) {
	claims, err := s.jwtMgr.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", nil, ErrSessionExpired
	}
	// 旧版本签发的 refresh token 没有 jti / fid，无法保证单次使用，要求重新登录。
	if claims.FamilyID == "" || claims.ID == "" {
		return "", "", nil, ErrSessionExpired
	}
	if s.sessions.IsRevoked(ctx, claims.FamilyID) {
		return "", "", nil, ErrSessionReused
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		_ = s.sessions.Revoke(ctx, claims.FamilyID)
		return "", "", nil, ErrSessionExpired
	}

	access, err := s.jwtMgr.IssueAccessToken(user.ID, user.Phone, claims.FamilyID)
	if err != nil {
		return "", "", nil, err
	}
	next, err := s.jwtMgr.IssueRefreshToken(user.ID, claims.FamilyID, claims.Generation+1)
	if err != nil {
		return "", "", nil, err
	}
	payload, err := json.Marshal(gracePayload{AccessToken: access, RefreshToken: next.Token})
	if err != nil {
		return "", "", nil, err
	}

	result, err := s.sessions.Rotate(ctx, claims.FamilyID, claims.ID, next.JTI, string(payload))
	if err != nil {
		return "", "", nil, err
	}
	switch result {
	case session.Rotated:
		return access, next.Token, user, nil
	case session.Replayed:
		// 同一旧 token 的并发刷新（多标签页 / 重试）：宽限窗口内返回同一对 token。
		if cached, ok := s.sessions.Grace(ctx, claims.ID); ok {
			var pair gracePayload
			if json.Unmarshal([]byte(cached), &pair) == nil && pair.AccessToken != "" && pair.RefreshToken != "" {
				return pair.AccessToken, pair.RefreshToken, user, nil
			}
		}
		// 宽限窗口外重放：认为 refresh token 已泄露，吊销整条家族链。
		if err := s.sessions.Revoke(ctx, claims.FamilyID); err != nil {
			slog.Error("revoke token family failed", "family_id", claims.FamilyID, "error", err)
		}
		return "", "", nil, ErrSessionReused
	case session.RevokedFlag:
		return "", "", nil, ErrSessionReused
	default:
		return "", "", nil, ErrSessionExpired
	}
}

// Logout 吊销 refresh token 所属家族（幂等：token 已失效也返回 nil）。
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.jwtMgr.ParseRefresh(refreshToken)
	if err != nil {
		return nil
	}
	if claims.FamilyID == "" {
		return nil
	}
	return s.sessions.Revoke(ctx, claims.FamilyID)
}

// IsSessionRevoked 供鉴权中间件检查 access token 所属家族是否已被吊销。
func (s *AuthService) IsSessionRevoked(ctx context.Context, familyID string) bool {
	return s.sessions.IsRevoked(ctx, familyID)
}

type gracePayload struct {
	AccessToken  string `json:"a"`
	RefreshToken string `json:"r"`
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func defaultNickname(phone string) string {
	if len(phone) >= 4 {
		return "用户" + phone[len(phone)-4:]
	}
	return "用户"
}
