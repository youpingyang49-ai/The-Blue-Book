package session

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// newTestStore 连接本地 Redis 的 15 号库；不可用时跳过，保证无 Redis 环境也能 go test。
func newTestStore(t *testing.T) (*Store, *redis.Client) {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Skipf("redis unavailable at %s: %v", addr, err)
	}
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush test db: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return NewStore(client, time.Hour), client
}

func TestRotateInvalidatesOldJTI(t *testing.T) {
	store, client := newTestStore(t)
	ctx := context.Background()
	const family = "f-rotate"

	if err := store.Start(ctx, family, "jti-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	result, err := store.Rotate(ctx, family, "jti-1", "jti-2", `{"a":"access-1"}`)
	if err != nil || result != Rotated {
		t.Fatalf("first rotate = %v, err=%v; want Rotated", result, err)
	}

	// 旧 jti 再提交一次：必须失败，且不是 Rotated。
	result, err = store.Rotate(ctx, family, "jti-1", "jti-3", `{"a":"access-2"}`)
	if err != nil {
		t.Fatalf("second rotate: %v", err)
	}
	if result != Replayed {
		t.Fatalf("old jti reuse = %v, want Replayed", result)
	}
	if current, _ := client.Get(ctx, familyKey(family)).Result(); current != "jti-2" {
		t.Fatalf("family current jti = %q, want jti-2", current)
	}

	// 新 jti 仍可正常轮换。
	result, err = store.Rotate(ctx, family, "jti-2", "jti-3", `{"a":"access-3"}`)
	if err != nil || result != Rotated {
		t.Fatalf("rotate with new jti = %v, err=%v; want Rotated", result, err)
	}
}

func TestConcurrentRefreshGetsSamePair(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	const family = "f-concurrent"
	const payload = `{"a":"access","r":"refresh"}`

	if err := store.Start(ctx, family, "jti-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	first, err := store.Rotate(ctx, family, "jti-1", "jti-2", payload)
	if err != nil || first != Rotated {
		t.Fatalf("first rotate = %v, err=%v", first, err)
	}
	second, err := store.Rotate(ctx, family, "jti-1", "jti-9", `{"a":"other"}`)
	if err != nil || second != Replayed {
		t.Fatalf("concurrent rotate = %v, err=%v; want Replayed", second, err)
	}
	cached, ok := store.Grace(ctx, "jti-1")
	if !ok || cached != payload {
		t.Fatalf("grace payload = %q ok=%v; want %q", cached, ok, payload)
	}
}

func TestRevokeBlocksFamily(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	const family = "f-revoke"

	if err := store.Start(ctx, family, "jti-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := store.Revoke(ctx, family); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !store.IsRevoked(ctx, family) {
		t.Fatal("family should be revoked")
	}
	result, err := store.Rotate(ctx, family, "jti-1", "jti-2", `{}`)
	if err != nil || result != RevokedFlag {
		t.Fatalf("rotate after revoke = %v, err=%v; want RevokedFlag", result, err)
	}
}

func TestMissingFamily(t *testing.T) {
	store, _ := newTestStore(t)
	result, err := store.Rotate(context.Background(), "f-missing", "jti-1", "jti-2", `{}`)
	if err != nil || result != Missing {
		t.Fatalf("rotate missing family = %v, err=%v; want Missing", result, err)
	}
	if store.IsRevoked(context.Background(), "f-missing") {
		t.Fatal("unknown family must not be considered revoked")
	}
}
