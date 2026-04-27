// Package config — загрузка настроек admin-bff из ENV.
package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ListenAddr      string
	IdentityURL     string
	BillingURL      string
	GameNovaURL     string
	RedisAddr       string
	SessionSecret   []byte
	IdleTimeout     time.Duration
	RefreshLeadTime time.Duration
	CookieDomain    string
	CookieSecure    bool
	LogLevel        string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:      env("BFF_LISTEN_ADDR", ":9200"),
		IdentityURL:     env("IDENTITY_URL", "http://localhost:9001"),
		BillingURL:      env("BILLING_URL", "http://localhost:9100"),
		GameNovaURL:     env("GAME_NOVA_URL", "http://localhost:8080"),
		RedisAddr:       env("REDIS_ADDR", "localhost:6379"),
		IdleTimeout:     envDuration("BFF_IDLE_TIMEOUT", 30*time.Minute),
		RefreshLeadTime: envDuration("BFF_REFRESH_LEAD_TIME", 60*time.Second),
		CookieDomain:    env("BFF_COOKIE_DOMAIN", ""),
		CookieSecure:    env("BFF_COOKIE_SECURE", "true") == "true",
		LogLevel:        env("LOG_LEVEL", "info"),
	}
	secret := env("SESSION_SECRET", "")
	if secret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required (32+ bytes hex/base64)")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET too short: %d bytes (need 32+)", len(secret))
	}
	cfg.SessionSecret = []byte(secret)
	return cfg, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
