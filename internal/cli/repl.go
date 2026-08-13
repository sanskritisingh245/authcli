package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"

	"authcli/internal/authsvc"
	"authcli/internal/models"
)

const prompt = "authcli> "

type REPL struct {
	svc   *authsvc.Service
	rl    *readline.Instance
	user  *models.User
	token string
}

func New(svc *authsvc.Service) (*REPL, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: prompt,
	})
	if err != nil {
		return nil, err
	}
	return &REPL{svc: svc, rl: rl}, nil
}

func (r *REPL) Close() error {
	return r.rl.Close()
}

func (r *REPL) Run(ctx context.Context) {
	fmt.Println("authcli — type 'help' for commands, 'exit' to quit")
	for {
		line, err := r.rl.Readline()
		if err == io.EOF || err == readline.ErrInterrupt {
			return
		}
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r.dispatch(ctx, line)
	}
}

func (r *REPL) dispatch(ctx context.Context, cmd string) {
	switch cmd {
	case "register":
		r.cmdRegister(ctx)
	case "login":
		r.cmdLogin(ctx)
	case "logout":
		r.cmdLogout(ctx)
	case "whoami":
		r.cmdWhoAmI(ctx)
	case "enable-2fa":
		r.cmdEnableTOTP(ctx)
	case "verify-2fa":
		r.cmdVerifyTOTP(ctx)
	case "disable-2fa":
		r.cmdDisableTOTP(ctx)
	case "help":
		r.cmdHelp()
	case "exit", "quit":
		r.Close()
		os.Exit(0)
	default:
		fmt.Println("unknown command, type 'help'")
	}
}
