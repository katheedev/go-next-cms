package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Port                string
	DatabaseURL         string
	SessionCookieName   string
	SessionCookieSecure bool
	UploadDir           string
	BaseURL             string
	MaxUploadBytes      int64
	DefaultLang         string
}

func Load() Config {
	_ = godotenv.Load()
	secure := getEnv("SESSION_COOKIE_SECURE", "false") == "true"
	maxUploadMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "5"), 10, 64)
	return Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "3000"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/deals?sslmode=disable"),
		SessionCookieName:   getEnv("SESSION_COOKIE_NAME", "deals_session"),
		SessionCookieSecure: secure,
		UploadDir:           getEnv("UPLOAD_DIR", "static/uploads"),
		BaseURL:             getEnv("BASE_URL", "http://localhost:3000"),
		MaxUploadBytes:      maxUploadMB * 1024 * 1024,
		DefaultLang:         getEnv("DEFAULT_LANG", "en"),
	}
}

func (c Config) Addr() string { return fmt.Sprintf(":%s", c.Port) }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
