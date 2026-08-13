package authsvc

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !checkPassword(hash, "correct horse battery staple") {
		t.Fatal("expected matching password to succeed")
	}
	if checkPassword(hash, "wrong password") {
		t.Fatal("expected wrong password to fail")
	}
}
