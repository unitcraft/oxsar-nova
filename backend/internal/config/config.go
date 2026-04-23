// Package config загружает настройки приложения из ENV и YAML-справочников.
//
// ENV — это изменяемые параметры окружения (адрес БД, секреты, флаги).
// YAML-справочники (здания, корабли, исследования) — источник игрового
// баланса; перезагружаются только при перезапуске процесса.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config агрегирует все настройки, необходимые приложению на старте.
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	Auth   AuthConfig
	Game   GameConfig
}

type ServerConfig struct {
	Addr     string
	Env      string
	LogLevel string
}

type DBConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type AuthConfig struct {
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	OAuthGoogleID string
	OAuthGoogleSc string
}

type GameConfig struct {
	Speed       float64
	Universe    string
	Deathmatch  bool
	NumGalaxies int
	NumSystems  int
	Points      PointsCoefficients
}

// PointsCoefficients — коэффициенты начисления очков.
// Значения по умолчанию соответствуют Dominator (consts.php).
type PointsCoefficients struct {
	Building float64 // очки = cost × k (за постройки)
	Research float64 // очки = cost × k (за исследования)
	Unit     float64 // очки = cost × k (за корабли/оборону)
}

// Load читает переменные окружения и возвращает валидированный Config.
// При отсутствии обязательного поля возвращает ошибку вместо panic,
// чтобы main мог логировать её с контекстом.
func Load() (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Addr:     env("SERVER_ADDR", ":8080"),
			Env:      env("SERVER_ENV", "dev"),
			LogLevel: env("LOG_LEVEL", "info"),
		},
		DB: DBConfig{
			URL: mustEnv("DB_URL"),
		},
		Redis: RedisConfig{
			URL: env("REDIS_URL", "redis://localhost:6379/0"),
		},
		Auth: AuthConfig{
			JWTSecret:     mustEnv("JWT_SECRET"),
			AccessTTL:     envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    envDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
			OAuthGoogleID: env("OAUTH_GOOGLE_CLIENT_ID", ""),
			OAuthGoogleSc: env("OAUTH_GOOGLE_CLIENT_SECRET", ""),
		},
		Game: GameConfig{
			Speed:       envFloat("GAMESPEED", 0.75),
			Universe:    env("UNIVERSE_NAME", "Nova"),
			Deathmatch:  envBool("DEATHMATCH", false),
			NumGalaxies: envInt("NUM_GALAXIES", 8),
			NumSystems:  envInt("NUM_SYSTEMS", 600),
			Points: PointsCoefficients{
				Building: envFloat("POINTS_K_BUILDING", 0.00005),
				Research: envFloat("POINTS_K_RESEARCH", 0.0005),
				Unit:     envFloat("POINTS_K_UNIT", 0.002),
			},
		},
	}

	if cfg.DB.URL == "" {
		return Config{}, fmt.Errorf("DB_URL is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return ""
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
