package main

import (
	"context"
	"log"

	"secrethole/backend/internal/app"
	"secrethole/backend/internal/config"
	"secrethole/backend/internal/db"
)

func main() {
	cfg := config.FromEnv()
	ctx := context.Background()
	d, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Pool.Close()

	s := app.NewServer(cfg, d)
	if err := s.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("API listening on", cfg.Addr)
	if err := s.Start(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
