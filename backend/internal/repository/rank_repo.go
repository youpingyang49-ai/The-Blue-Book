package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RankRepo 封装排行榜用到的 Redis ZSet 操作。
// 只使用 Redis 5 也支持的指令（ZADD / ZREVRANGE / ZREMRANGEBYRANK / RENAME），
// 便于本地便携版与容器版行为一致。
type RankRepo struct {
	rdb *redis.Client
}

func NewRankRepo(rdb *redis.Client) *RankRepo {
	return &RankRepo{rdb: rdb}
}

// SetScores 批量写入 member -> score。
func (r *RankRepo) SetScores(ctx context.Context, key string, scores map[string]float64) error {
	if len(scores) == 0 {
		return nil
	}
	members := make([]redis.Z, 0, len(scores))
	for member, score := range scores {
		members = append(members, redis.Z{Score: score, Member: member})
	}
	return r.rdb.ZAdd(ctx, key, members...).Err()
}

// Top 取榜单第 offset 名起的 limit 条（score 从高到低）。
func (r *RankRepo) Top(ctx context.Context, key string, offset, limit int64) ([]redis.Z, error) {
	if limit <= 0 {
		return nil, nil
	}
	return r.rdb.ZRevRangeWithScores(ctx, key, offset, offset+limit-1).Result()
}

// Card 返回榜单大小。
func (r *RankRepo) Card(ctx context.Context, key string) (int64, error) {
	return r.rdb.ZCard(ctx, key).Result()
}

// Remove 从榜单移除若干 member（笔记删除时调用）。
func (r *RankRepo) Remove(ctx context.Context, keys []string, member string) error {
	if len(keys) == 0 {
		return nil
	}
	pipe := r.rdb.Pipeline()
	for _, key := range keys {
		pipe.ZRem(ctx, key, member)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Trim 只保留分数最高的 max 条，避免榜单无限增长。
func (r *RankRepo) Trim(ctx context.Context, key string, max int64) error {
	if max <= 0 {
		return nil
	}
	return r.rdb.ZRemRangeByRank(ctx, key, 0, -(max + 1)).Err()
}

// Replace 用临时键原子替换正式键；临时键不存在时清空正式键（空榜）。
func (r *RankRepo) Replace(ctx context.Context, tmpKey, liveKey string) error {
	if err := r.rdb.Rename(ctx, tmpKey, liveKey).Err(); err == nil {
		return nil
	}
	return r.rdb.Del(ctx, liveKey).Err()
}
