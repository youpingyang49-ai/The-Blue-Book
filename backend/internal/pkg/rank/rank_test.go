package rank

import (
	"math"
	"testing"
	"time"
)

func TestHeatPrefersFreshAndPopular(t *testing.T) {
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	base := Metrics{Likes: 10, Favorites: 5, Comments: 3, CreatedAt: now.Add(-2 * time.Hour)}
	fresh := Metrics{Likes: 10, Favorites: 5, Comments: 3, CreatedAt: now.Add(-1 * time.Hour)}
	popular := Metrics{Likes: 100, Favorites: 50, Comments: 30, CreatedAt: now.Add(-2 * time.Hour)}

	if Heat(fresh, now) <= Heat(base, now) {
		t.Fatalf("newer post should rank higher: fresh=%v base=%v", Heat(fresh, now), Heat(base, now))
	}
	if Heat(popular, now) <= Heat(base, now) {
		t.Fatalf("more engagement should rank higher: popular=%v base=%v", Heat(popular, now), Heat(base, now))
	}
	if got := Heat(Metrics{CreatedAt: now}, now); got <= 0 || math.IsInf(got, 0) || math.IsNaN(got) {
		t.Fatalf("zero-engagement post should still have a positive finite score, got %v", got)
	}
}

func TestHeatClampsFuturePosts(t *testing.T) {
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	future := Metrics{Likes: 1, CreatedAt: now.Add(48 * time.Hour)}
	current := Metrics{Likes: 1, CreatedAt: now}
	if Heat(future, now) != Heat(current, now) {
		t.Fatalf("future timestamp should clamp to age 0: future=%v current=%v", Heat(future, now), Heat(current, now))
	}
}

func TestMemberRoundTripAndOrder(t *testing.T) {
	for _, id := range []uint{1, 42, 999999} {
		member := PostMember(id)
		got, ok := ParseMember(member)
		if !ok || got != id {
			t.Fatalf("round trip failed for %d: member=%q got=%d ok=%v", id, member, got, ok)
		}
	}
	if PostMember(2) >= PostMember(10) {
		t.Fatalf("members must sort numerically: %q >= %q", PostMember(2), PostMember(10))
	}
	if _, ok := ParseMember(""); ok {
		t.Fatal("empty member should be rejected")
	}
	if _, ok := ParseMember("abc"); ok {
		t.Fatal("non numeric member should be rejected")
	}
}

func TestNormalize(t *testing.T) {
	if got := Normalize(0.5, 1); got != 50 {
		t.Fatalf("normalize 0.5/1 = %d, want 50", got)
	}
	if got := Normalize(1, 1); got != 100 {
		t.Fatalf("top score should normalize to 100, got %d", got)
	}
	if got := Normalize(0.000001, 1); got != 1 {
		t.Fatalf("tiny score should normalize to at least 1, got %d", got)
	}
	if got := Normalize(1, 0); got != 0 {
		t.Fatalf("zero top should normalize to 0, got %d", got)
	}
}
