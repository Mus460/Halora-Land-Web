package main

import (
	"context"
	"log"

	"github.com/halora-land/halora-be/internal/auth"
)

// initAuth wires the local-auth verifier and ensures the bootstrap admin
// account exists on a fresh database.
func (app *App) initAuth(ctx context.Context) {
	app.verifier = auth.NewVerifier(app.pool, app.cfg.JWTSecret)

	created, err := app.verifier.EnsureDefaultAdmin(ctx)
	if err != nil {
		log.Printf("WARNING: ensure default admin: %v", err)
		return
	}
	if created {
		log.Printf("bootstrap admin created: %s / admin123", auth.DefaultAdminEmail)
	}
}
