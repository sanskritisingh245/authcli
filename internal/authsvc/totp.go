package authsvc

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func generateTOTPSecret(issuer, username string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
}

func validateTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}
