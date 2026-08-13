package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	LockoutMaxAttempts     int
	LockoutDurationMinutes int
	SessionDurationHours   int
	TOTPIssuer             string
}

func Load() (Config, error) {
	cfg := Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "authcli"),
		DBPassword: getEnv("DB_PASSWORD", "authcli"),
		DBName:     getEnv("DB_NAME", "authcli"),
		TOTPIssuer: getEnv("TOTP_ISSUER", "AuthCLI"),
	}

	var err error
	if cfg.LockoutMaxAttempts, err = getEnvInt("LOCKOUT_MAX_ATTEMPTS", 5); err != nil {
		return Config{}, err
	}
	if cfg.LockoutDurationMinutes, err = getEnvInt("LOCKOUT_DURATION_MINUTES", 15); err != nil {
		return Config{}, err
	}
	if cfg.SessionDurationHours, err = getEnvInt("SESSION_DURATION_HOURS", 24); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword), c.DBHost, c.DBPort, c.DBName)
}

func (c Config) LockoutDuration() time.Duration {
	return time.Duration(c.LockoutDurationMinutes) * time.Minute
}

func (c Config) SessionDuration() time.Duration {
	return time.Duration(c.SessionDurationHours) * time.Hour
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	return strconv.Atoi(v)
}
