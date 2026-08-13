package authsvc

import "testing"

func TestGenerateTokenUnique(t *testing.T) {
	a, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected two generated tokens to differ")
	}
	if len(a) != 64 {
		t.Fatalf("expected 64 hex chars (32 bytes), got %d", len(a))
	}
}

func TestHashToken(t *testing.T) {
	if hashToken("abc") != hashToken("abc") {
		t.Fatal("expected hash to be deterministic")
	}
	if hashToken("abc") == hashToken("abd") {
		t.Fatal("expected different tokens to hash differently")
	}
}
