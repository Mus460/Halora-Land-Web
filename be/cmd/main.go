package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/env"
	"github.com/halora-land/halora-be/internal/ratelimit"
)

const DEFAULT_AHSP_PATH = "./data/ahsp-2026.xlsx"

// App holds the long-lived dependencies shared by every route.
type App struct {
	cfg      *env.Config
	pool     *pgxpool.Pool
	verifier *auth.Verifier
	audit    *audit.Logger
	limiter  *ratelimit.Limiter
	ahspPath string
}

func main() {
	_ = godotenv.Load() // best-effort: ignore error if no .env

	decimal.MarshalJSONWithoutQuotes = true // serialize decimal.Decimal as JSON numbers, not strings

	if len(os.Args) < 2 {
		fmt.Println("usage: halora-be <serve|migrate|import-ahsp|seed> [flags]")
		os.Exit(1)
	}

	cfg, err := env.Load()
	if err != nil {
		log.Fatal(err)
	}

	cmd(os.Args[1], cfg)
}

func cmd(input string, cfg *env.Config) {
	switch input {
	case "serve":
		if err := serve(cfg); err != nil {
			log.Fatal(err)
		}
	case "migrate":
		if err := migrate(cfg); err != nil {
			log.Fatal(err)
		}
	case "import-ahsp":
		if err := importAHSP(cfg); err != nil {
			log.Fatal(err)
		}
	case "seed":
		if err := seed(cfg); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(1)
	}

}

// serve builds the App and runs the HTTP server until SIGINT/SIGTERM.
func serve(cfg *env.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := &App{cfg: cfg, ahspPath: DEFAULT_AHSP_PATH}

	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	app.pool = pool

	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if cfg.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	app.verifier = auth.NewVerifier(pool, cfg.JWTSecret)
	app.limiter = ratelimit.New()
	app.audit = audit.New(pool, 1024)
	defer app.audit.Close()

	app.initAuth(ctx)

	return app.serve(ctx)
}
