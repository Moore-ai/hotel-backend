package hash

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mysecret")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "mysecret" {
		t.Fatal("password should not be stored in plaintext")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, _ := HashPassword("mysecret")
	if !CheckPassword(hash, "mysecret") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected wrong password to not match")
	}
}
