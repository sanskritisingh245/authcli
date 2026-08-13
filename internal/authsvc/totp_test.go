package authsvc

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateAndValidateTOTP(t *testing.T) {
	key, err := generateTOTPSecret("AuthCLI", "alice")
	if err != nil {
		t.Fatal(err)
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !validateTOTPCode(key.Secret(), code) {
		t.Fatal("expected freshly generated code to validate")
	}
	if validateTOTPCode(key.Secret(), "000000") {
		t.Fatal("expected an arbitrary wrong code to fail")
	}
}
