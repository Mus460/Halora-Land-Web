package main

import (
	"context"
	"log"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/env"
)

func migrate(cfg *env.Config) error {
	ctx := context.Background()
	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	log.Println("migrations applied")
	return nil
}
