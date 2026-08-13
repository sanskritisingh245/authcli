package main

import (
	"context"
	"log"

	"authcli/internal/authsvc"
	"authcli/internal/cli"
	"authcli/internal/config"
	"authcli/internal/db"
	"authcli/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(cfg.DSN(), "migrations"); err != nil {
		log.Fatal(err)
	}

	pool, err := db.Connect(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	svc := authsvc.New(store.New(pool), cfg)

	repl, err := cli.New(svc)
	if err != nil {
		log.Fatal(err)
	}
	defer repl.Close()

	repl.Run(ctx)
}
