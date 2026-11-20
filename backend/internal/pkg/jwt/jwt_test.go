package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenTypesAndFamilyMetadata(t *testing.T) {
	manager := NewManager("test-secret", time.Minute, time.Hour)
	familyID, err := NewFamilyID()
	if err != nil {
		t.Fatalf("new family id: %v", err)
	}
	refresh, err := manager.IssueRefreshToken(7, familyID, 2)
	if err != nil {
		t.Fatalf("issue refresh: %v", err)
	}
	claims, err := manager.ParseRefresh(refresh.Token)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if claims.UserID != 7 || claims.FamilyID != familyID || claims.Generation != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ID != refresh.JTI || claims.ID == "" {
		t.Fatalf("jti mismatch: claims=%q issued=%q", claims.ID, refresh.JTI)
	}

	access, err := manager.IssueAccessToken(7, "13800000000", familyID)
	if err != nil {
		t.Fatalf("issue access: %v", err)
	}
	if _, err := manager.ParseRefresh(access); err == nil {
		t.Fatal("access token must not parse as refresh token")
	}
	if _, err := manager.ParseAccess(refresh.Token); err == nil {
		t.Fatal("refresh token must not parse as access token")
	}
}

func TestRejectsExpiredAndUnsignedTokens(t *testing.T) {
	manager := NewManager("test-secret", time.Minute, time.Hour)

	expired := Claims{
		UserID: 1,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expired).SignedString(manager.secret)
	if err != nil {
		t.Fatalf("sign expired: %v", err)
	}
	if _, err := manager.ParseAccess(raw); err == nil {
		t.Fatal("expired token should be rejected")
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: 1, Type: TokenTypeAccess})
	rawNone, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	if _, err := manager.ParseAccess(rawNone); err == nil {
		t.Fatal("unsigned token should be rejected")
	}
}

func TestTTLDefaults(t *testing.T) {
	manager := NewManager("secret", 0, 0)
	if manager.AccessTTL() != DefaultAccessTTL || manager.RefreshTTL() != DefaultRefreshTTL {
		t.Fatalf("unexpected defaults: %v / %v", manager.AccessTTL(), manager.RefreshTTL())
	}
}
