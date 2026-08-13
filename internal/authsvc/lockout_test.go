package authsvc

import (
	"testing"
	"time"

	"authcli/internal/models"
)

func TestIsLocked(t *testing.T) {
	future := time.Now().Add(time.Minute)
	past := time.Now().Add(-time.Minute)

	cases := []struct {
		name string
		user models.User
		want bool
	}{
		{"no lock set", models.User{}, false},
		{"locked until future", models.User{LockedUntil: &future}, true},
		{"lock already expired", models.User{LockedUntil: &past}, false},
	}

	for _, c := range cases {
		if got := isLocked(c.user); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
