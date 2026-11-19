package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

var (
	ErrCodeExpired    = errors.New("验证码已过期")
	ErrCodeInvalid    = errors.New("验证码错误")
	ErrRateLimited    = errors.New("发送太频繁，请稍后再试")
	ErrSessionExpired = errors.New("登录状态已失效")
)

const (
	codeTTL = 5 * time.Minute
	rateTTL = 60 * time.Second
)

type AuthService struct {
	userRepo *repository.UserRepo
	rdb      *redis.Client
	jwtMgr   *jwt.Manager
}

func NewAuthService(userRepo *repository.UserRepo, rdb *redis.Client, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{userRepo: userRepo, rdb: rdb, jwtMgr: jwtMgr}
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

	access, err := s.jwtMgr.GenerateAccessToken(user.ID, user.Phone)
	if err != nil {
		return "", "", nil, err
	}
	refresh, err := s.jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}
	return access, refresh, user, nil
}

// Refresh 用 refresh token 换取新的 access / refresh token（刷新令牌轮换）。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, *model.User, error) {
	claims, err := s.jwtMgr.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", nil, ErrSessionExpired
	}
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		return "", "", nil, ErrSessionExpired
	}
	access, err := s.jwtMgr.GenerateAccessToken(user.ID, user.Phone)
	if err != nil {
		return "", "", nil, err
	}
	refresh, err := s.jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}
	return access, refresh, user, nil
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
