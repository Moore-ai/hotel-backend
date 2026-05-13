package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour

	token, err := GenerateToken(1, "admin", secret, expiry)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Fatalf("expected userID 1, got %d", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Fatalf("expected role admin, got %s", claims.Role)
	}
}

func TestParseExpiredToken(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateToken(1, "guest", secret, -time.Hour)
	_, err := ParseToken(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}
