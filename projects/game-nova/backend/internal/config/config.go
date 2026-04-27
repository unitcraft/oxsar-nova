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
//
// План 38 Ф.5: Payment удалён — платежи в billing-service (отдельный домен).
type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Game      GameConfig
	AIAdvisor AIAdvisorConfig
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
	OAuthGoogleID string
	OAuthGoogleSc string
	// JWKSUrl — URL Auth Service: http://auth-service:9000/.well-known/jwks.json
	JWKSUrl        string
	// AuthServiceURL — базовый URL Auth Service для межсервисных вызовов.
	// Пример: http://auth-service:9000
	AuthServiceURL string
	// UniverseID — идентификатор этой вселенной (uni01, uni02, …).
	UniverseID string
}

type GameConfig struct {
	Speed                  float64
	Universe               string
	Deathmatch             bool
	NumGalaxies            int
	NumSystems             int
	Points                 PointsCoefficients
	StorageFactor          float64 // STORAGE_FACTOR (Dominator=5)
	ResearchSpeedFactor    float64 // RESEARCH_SPEED_FACTOR (Dominator=2)
	EnergyProductionFactor float64 // ENEGRY_PRODUCTION_FACTOR (Dominator=0.8)
	MaxPlanets             int     // MAX_PLANETS per player, 0 = computer_tech+1
	BashingPeriod          int     // seconds, 0 = disabled
	BashingMaxAttacks      int     // max attacks per BashingPeriod
	ProtectionPeriod       int     // seconds new player is protected from attacks
}

type AIAdvisorConfig struct {
	APIKey      string // ANTHROPIC_API_KEY
	ProxyURL    string // ANTHROPIC_PROXY_URL (опционально)
	OllamaURL   string // OLLAMA_URL (если задан — использовать Ollama вместо Claude)
	OllamaModel string // OLLAMA_MODEL, default "qwen2.5:3b"
	MaxPerDay   int    // AI_ADVISOR_MAX_PER_DAY, default 20
	MaxTokens   int    // AI_ADVISOR_MAX_TOKENS, default 1024
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
			OAuthGoogleID:  env("OAUTH_GOOGLE_CLIENT_ID", ""),
			OAuthGoogleSc:  env("OAUTH_GOOGLE_CLIENT_SECRET", ""),
			JWKSUrl:        env("AUTH_JWKS_URL", ""),
			AuthServiceURL: env("AUTH_SERVICE_URL", ""),
			UniverseID:     env("UNIVERSE_ID", "uni01"),
		},
		Game: GameConfig{
			Speed:                  envFloat("GAMESPEED", 0.75),
			Universe:               env("UNIVERSE_NAME", "Nova"),
			Deathmatch:             envBool("DEATHMATCH", false),
			NumGalaxies:            envInt("NUM_GALAXIES", 8),
			NumSystems:             envInt("NUM_SYSTEMS", 600),
			StorageFactor:          envFloat("STORAGE_FACTOR", 5),
			ResearchSpeedFactor:    envFloat("RESEARCH_SPEED_FACTOR", 2),
			EnergyProductionFactor: envFloat("ENERGY_PRODUCTION_FACTOR", 0.8),
			MaxPlanets:             envInt("MAX_PLANETS", 0),
			BashingPeriod:          envInt("BASHING_PERIOD", 18000),
			BashingMaxAttacks:      envInt("BASHING_MAX_ATTACKS", 4),
			ProtectionPeriod:       envInt("PROTECTION_PERIOD", 86400),
			Points: PointsCoefficients{
				Building: envFloat("POINTS_K_BUILDING", 0.00005),
				Research: envFloat("POINTS_K_RESEARCH", 0.0005),
				Unit:     envFloat("POINTS_K_UNIT", 0.002),
			},
		},
	}

	cfg.AIAdvisor = AIAdvisorConfig{
		APIKey:      env("ANTHROPIC_API_KEY", ""),
		ProxyURL:    env("ANTHROPIC_PROXY_URL", ""),
		OllamaURL:   env("OLLAMA_URL", ""),
		OllamaModel: env("OLLAMA_MODEL", "qwen2.5:3b"),
		MaxPerDay:   envInt("AI_ADVISOR_MAX_PER_DAY", 20),
		MaxTokens:   envInt("AI_ADVISOR_MAX_TOKENS", 1024),
	}

	// План 38 Ф.5: PaymentConfig удалён — платежи в billing-service.
	// PAYMENT_* env-переменные читаются billing-service-ом, не game-nova.

	if cfg.DB.URL == "" {
		return Config{}, fmt.Errorf("DB_URL is required")
	}
	// План 36 Ф.12: единственный режим — RSA через JWKS (AUTH_JWKS_URL обязателен).
	// HS256 fallback с JWT_SECRET удалён.
	if cfg.Auth.JWKSUrl == "" {
		return Config{}, fmt.Errorf("AUTH_JWKS_URL is required")
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
