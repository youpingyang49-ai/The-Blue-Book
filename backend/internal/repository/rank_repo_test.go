package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func newTestRankRepo(t *testing.T) *RankRepo {
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
	t.Cleanup(func() { _ = client.Close() })
	return NewRankRepo(client)
}

func TestZSetRankingReadWrite(t *testing.T) {
	repo := newTestRankRepo(t)
	ctx := context.Background()
	key := "test:rank:posts"
	t.Cleanup(func() { repo.rdb.Del(ctx, key, key+":tmp") })

	if err := repo.SetScores(ctx, key, map[string]float64{
		"000000000001": 0.5,
		"000000000002": 2.5,
		"000000000003": 1.5,
	}); err != nil {
		t.Fatalf("set scores: %v", err)
	}
	entries, err := repo.Top(ctx, key, 0, 2)
	if err != nil {
		t.Fatalf("top: %v", err)
	}
	if len(entries) != 2 || entries[0].Member != "000000000002" || entries[1].Member != "000000000003" {
		t.Fatalf("unexpected top order: %+v", entries)
	}
	if card, _ := repo.Card(ctx, key); card != 3 {
		t.Fatalf("card = %d, want 3", card)
	}
	if err := repo.Trim(ctx, key, 2); err != nil {
		t.Fatalf("trim: %v", err)
	}
	if card, _ := repo.Card(ctx, key); card != 2 {
		t.Fatalf("card after trim = %d, want 2", card)
	}
	if err := repo.Remove(ctx, []string{key}, "000000000003"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if card, _ := repo.Card(ctx, key); card != 1 {
		t.Fatalf("card after remove = %d, want 1", card)
	}
}

func TestZSetReplaceSwapsAtomically(t *testing.T) {
	repo := newTestRankRepo(t)
	ctx := context.Background()
	live, tmp := "test:rank:topics", "test:rank:topics:tmp"
	t.Cleanup(func() { repo.rdb.Del(ctx, live, tmp) })

	if err := repo.SetScores(ctx, live, map[string]float64{"1": 1}); err != nil {
		t.Fatalf("seed live: %v", err)
	}
	if err := repo.SetScores(ctx, tmp, map[string]float64{"2": 2, "3": 3}); err != nil {
		t.Fatalf("seed tmp: %v", err)
	}
	if err := repo.Replace(ctx, tmp, live); err != nil {
		t.Fatalf("replace: %v", err)
	}
	entries, _ := repo.Top(ctx, live, 0, 10)
	if len(entries) != 2 || entries[0].Member != "3" {
		t.Fatalf("unexpected live board after replace: %+v", entries)
	}

	// 空榜替换：临时键不存在时清空正式键。
	if err := repo.Replace(ctx, tmp, live); err != nil {
		t.Fatalf("replace empty: %v", err)
	}
	if card, _ := repo.Card(ctx, live); card != 0 {
		t.Fatalf("live board should be cleared, card=%d", card)
	}
}
