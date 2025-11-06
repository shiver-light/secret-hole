package config

import (
	"log"
	"os"
)

type Config struct {
	Addr        string // :8080
	DatabaseURL string // postgres://user:pass@localhost:5432/shudong?sslmode=disable
	Env         string // dev|prod
	CorsOrigins string // *, or https://your.app
}

func FromEnv() Config {
	cfg := Config{
		Addr:        getOr("ADDR", ":8080"),
		DatabaseURL: getOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shudong?sslmode=disable"),
		Env:         getOr("ENV", "dev"),
		CorsOrigins: getOr("CORS_ORIGINS", "*"),
	}
	log.Printf("cfg: %+v\n", cfg)
	return cfg
}

func getOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
