# authcli

Interactive CLI auth system in Go, backed by Postgres, with bcrypt password
hashing, TOTP 2FA, and session-based login.

## Architecture

![Architecture diagram](docs/architecture.png)

`cli` is the only layer that touches a terminal. `authsvc` is the only layer
that makes a security decision. `store` is the only layer that writes SQL.
Solid arrows are calls; dashed arrows are what comes back (including the
mid-`login` round trip back to the terminal for a 2FA code).

## Run it

```
cp .env.example .env
docker compose run --rm app
```

`docker compose run` (not `up`) because the app is an interactive shell and
needs a real stdin/tty attached. It builds the image and starts Postgres
automatically on first run.

## Commands

Once inside the `authcli>` prompt:

```
register     create a new account
login        log in (prompts for a 2fa code if enabled)
logout       invalidate the current session
whoami       show the logged in user + last login time
enable-2fa   generate a totp secret + qr code
verify-2fa   confirm setup with a 6-digit code
disable-2fa  turn off 2fa (requires a current code)
help         show this message
exit         quit
```

## Config (env vars, see `.env.example`)

| Var | Default | Meaning |
|---|---|---|
| `LOCKOUT_MAX_ATTEMPTS` | 5 | failed attempts before an account locks |
| `LOCKOUT_DURATION_MINUTES` | 15 | lockout cooldown |
| `SESSION_DURATION_HOURS` | 24 | session token lifetime |
| `TOTP_ISSUER` | AuthCLI | name shown in the authenticator app |

## Local dev without Docker

Needs a local Postgres reachable via `DB_HOST`/`DB_PORT`/etc. (defaults to
`localhost:5432`). Migrations run automatically on startup.

```
go run ./cmd/authcli
```

## Tests

```
go test ./...
```
