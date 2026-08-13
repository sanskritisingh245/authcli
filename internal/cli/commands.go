package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

func (r *REPL) readLine(p string) (string, error) {
	r.rl.SetPrompt(p)
	defer r.rl.SetPrompt(prompt)
	line, err := r.rl.Readline()
	return strings.TrimSpace(line), err
}

func (r *REPL) readPassword(p string) (string, error) {
	b, err := r.rl.ReadPassword(p)
	return string(b), err
}

func (r *REPL) cmdRegister(ctx context.Context) {
	username, err := r.readLine("Username: ")
	if err != nil {
		return
	}
	password, err := r.readPassword("Password: ")
	if err != nil {
		return
	}
	confirm, err := r.readPassword("Confirm:  ")
	if err != nil {
		return
	}
	if password != confirm {
		fmt.Println("passwords do not match")
		return
	}
	if err := r.svc.Register(ctx, username, password); err != nil {
		fmt.Println("registration failed:", err)
		return
	}
	fmt.Println("registered. run 'login' to continue.")
}

func (r *REPL) cmdLogin(ctx context.Context) {
	if r.user != nil {
		fmt.Println("already logged in as", r.user.Username)
		return
	}
	username, err := r.readLine("Username: ")
	if err != nil {
		return
	}
	password, err := r.readPassword("Password: ")
	if err != nil {
		return
	}

	promptTOTP := func() (string, error) {
		return r.readLine("2FA code: ")
	}

	user, token, err := r.svc.Login(ctx, username, password, promptTOTP)
	if err != nil {
		fmt.Println("login failed:", err)
		return
	}
	r.user = &user
	r.token = token
	fmt.Println("logged in as", user.Username)
}

func (r *REPL) cmdLogout(ctx context.Context) {
	if r.user == nil {
		fmt.Println("not logged in")
		return
	}
	if err := r.svc.Logout(ctx, r.token); err != nil {
		fmt.Println("logout failed:", err)
		return
	}
	r.user = nil
	r.token = ""
	fmt.Println("logged out")
}

func (r *REPL) cmdWhoAmI(ctx context.Context) {
	if r.user == nil {
		fmt.Println("not logged in")
		return
	}
	user, err := r.svc.ValidateSession(ctx, r.token)
	if err != nil {
		fmt.Println("session no longer valid:", err)
		r.user = nil
		r.token = ""
		return
	}
	r.user = &user

	lastLogin := "never"
	if user.LastLoginAt != nil {
		lastLogin = user.LastLoginAt.Format(time.RFC3339)
	}
	fmt.Printf("username:    %s\nlast login:  %s\n", user.Username, lastLogin)
}

func (r *REPL) cmdEnableTOTP(ctx context.Context) {
	if r.user == nil {
		fmt.Println("not logged in")
		return
	}
	key, err := r.svc.EnableTOTP(ctx, r.user.ID, r.user.Username)
	if err != nil {
		fmt.Println("could not enable 2fa:", err)
		return
	}

	q, err := qrcode.New(key.URL(), qrcode.Medium)
	if err == nil {
		fmt.Println(q.ToString(false))
	}
	fmt.Println("secret:", key.Secret())
	fmt.Println("otpauth uri:", key.URL())
	fmt.Println("run 'verify-2fa' and enter the 6-digit code to activate")
}

func (r *REPL) cmdVerifyTOTP(ctx context.Context) {
	if r.user == nil {
		fmt.Println("not logged in")
		return
	}
	code, err := r.readLine("6-digit code: ")
	if err != nil {
		return
	}
	if err := r.svc.ConfirmTOTP(ctx, r.user.ID, code); err != nil {
		fmt.Println("verification failed:", err)
		return
	}
	fmt.Println("2fa enabled")
}

func (r *REPL) cmdDisableTOTP(ctx context.Context) {
	if r.user == nil {
		fmt.Println("not logged in")
		return
	}
	code, err := r.readLine("current 6-digit code: ")
	if err != nil {
		return
	}
	if err := r.svc.DisableTOTP(ctx, r.user.ID, code); err != nil {
		fmt.Println("could not disable 2fa:", err)
		return
	}
	fmt.Println("2fa disabled")
}

func (r *REPL) cmdHelp() {
	fmt.Println(`commands:
  register     create a new account
  login        log in (prompts for 2fa code if enabled)
  logout       invalidate the current session
  whoami       show the logged in user
  enable-2fa   generate a totp secret + qr code
  verify-2fa   confirm setup with a 6-digit code
  disable-2fa  turn off 2fa (requires current code)
  help         show this message
  exit         quit`)
}
