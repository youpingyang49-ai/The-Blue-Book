// Package session 维护 refresh token 的「令牌家族」状态。
//
// 每次登录创建一个家族（family），家族内只允许当前一代 refresh token 使用一次：
//   - 轮换成功：旧 jti 立刻失效，新 jti 成为家族当前值；
//   - 并发刷新：旧 jti 在宽限窗口内重放时返回同一对 token（幂等，避免多标签页互相踢下线）；
//   - 宽限期外重放：判定为凭证泄露，吊销整个家族。
//
// 状态全部存在 Redis（兼容 Redis 5：仅用 SET EX / GET / EXISTS / DEL / Lua），
// 因此重启进程不会丢失单次使用语义。
package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "bluebook:auth:"
	// GraceTTL 轮换宽限窗口：同一旧 token 在窗口内重复提交视为并发刷新。
	GraceTTL = 60 * time.Second
	// ttlSlack 让家族状态比 JWT 本身多活一点，避免边界上「token 未过期但状态已丢」。
	ttlSlack = time.Minute
)

// Result 是一次轮换尝试的结果。
type Result int

const (
	// Rotated 轮换成功。
	Rotated Result = iota
	// Replayed 旧 jti 与家族当前值不一致：可能是并发刷新，也可能是重放，需查宽限缓存。
	Replayed
	// RevokedFlag 家族已被吊销。
	RevokedFlag
	// Missing 家族不存在或已过期。
	Missing
)

// Store 是令牌家族的 Redis 存储。
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewStore(rdb *redis.Client, refreshTTL time.Duration) *Store {
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &Store{rdb: rdb, ttl: refreshTTL + ttlSlack}
}

// Start 登录时登记家族当前有效的 refresh jti。
func (s *Store) Start(ctx context.Context, familyID, jti string) error {
	if familyID == "" || jti == "" {
		return errors.New("session: family id and jti are required")
	}
	return s.rdb.Set(ctx, familyKey(familyID), jti, s.ttl).Err()
}

// Rotate 原子地校验旧 jti 并替换为新 jti。
// 成功时把新的 token 对（gracePayload）写入旧 jti 的宽限键，供并发刷新复用。
func (s *Store) Rotate(ctx context.Context, familyID, oldJTI, newJTI, gracePayload string) (Result, error) {
	res, err := rotateScript.Run(ctx, s.rdb,
		[]string{familyKey(familyID), revokedKey(familyID), graceKey(oldJTI)},
		oldJTI, newJTI, int(s.ttl.Seconds()), gracePayload, int(GraceTTL.Seconds()),
	).Int()
	if err != nil {
		return Missing, err
	}
	switch res {
	case 1:
		return Rotated, nil
	case -1:
		return RevokedFlag, nil
	case -2:
		return Missing, nil
	case -3:
		return Replayed, nil
	default:
		return Missing, errors.New("session: unexpected rotate result")
	}
}

// Grace 返回宽限窗口内缓存的 token 对。
func (s *Store) Grace(ctx context.Context, oldJTI string) (string, bool) {
	if oldJTI == "" {
		return "", false
	}
	payload, err := s.rdb.Get(ctx, graceKey(oldJTI)).Result()
	if err != nil {
		return "", false
	}
	return payload, true
}

// Revoke 吊销整个家族（登出、检测到重放、用户不存在）。
func (s *Store) Revoke(ctx context.Context, familyID string) error {
	if familyID == "" {
		return nil
	}
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, revokedKey(familyID), "1", s.ttl)
	pipe.Del(ctx, familyKey(familyID))
	_, err := pipe.Exec(ctx)
	return err
}

// IsRevoked 查询家族是否已被吊销。
// Redis 故障时返回 false（fail-open）：可用性优先，access token 本身仍是短 TTL。
func (s *Store) IsRevoked(ctx context.Context, familyID string) bool {
	if familyID == "" {
		return false
	}
	n, err := s.rdb.Exists(ctx, revokedKey(familyID)).Result()
	if err != nil {
		slog.Warn("session revoke check failed", "error", err)
		return false
	}
	return n > 0
}

func familyKey(familyID string) string  { return keyPrefix + "family:" + familyID }
func revokedKey(familyID string) string { return keyPrefix + "family:revoked:" + familyID }
func graceKey(jti string) string        { return keyPrefix + "grace:" + jti }

// rotateScript 保证「校验 + 替换 + 写宽限缓存」是一个原子操作。
// KEYS: 1=家族当前 jti, 2=吊销标记, 3=旧 jti 的宽限缓存
// ARGV: 1=旧 jti, 2=新 jti, 3=家族 TTL 秒, 4=宽限内容, 5=宽限 TTL 秒
var rotateScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[2]) == 1 then
  return -1
end
local current = redis.call('GET', KEYS[1])
if not current then
  return -2
end
if current ~= ARGV[1] then
  return -3
end
redis.call('SET', KEYS[1], ARGV[2], 'EX', ARGV[3])
redis.call('SET', KEYS[3], ARGV[4], 'EX', ARGV[5])
return 1
`)
