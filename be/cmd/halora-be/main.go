package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/ahsp"
	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/config"
	"github.com/halora-land/halora-be/internal/db"
	"github.com/halora-land/halora-be/internal/migration"
	"github.com/halora-land/halora-be/internal/ratelimit"
	"github.com/halora-land/halora-be/internal/server"
)

const defaultAHSPPath = "./data/ahsp-2026.xlsx"

func main() {
	_ = godotenv.Load() // best-effort: ignore error if no .env

	decimal.MarshalJSONWithoutQuotes = true // serialize decimal.Decimal as JSON numbers, not strings

	if len(os.Args) < 2 {
		fmt.Println("usage: halora-be <serve|migrate|import-ahsp|seed> [flags]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "serve":
		serve(cfg)
	case "migrate":
		migrate(cfg)
	case "import-ahsp":
		importAHSP(cfg)
	case "seed":
		seed(cfg)
	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func serve(cfg *config.Config) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := migration.Run(ctx, pool); err != nil {
		log.Fatal("migrate: ", err)
	}

	limiter := ratelimit.New()

	if cfg.DemoMode {
		log.Println("WARNING: DEMO MODE enabled — using local DB auth (no Supabase)")
	}

	verifier := auth.NewVerifier(pool, cfg.SupabaseJWKSURL, cfg.SupabaseAnonKey, cfg.SupabaseProjectRef, cfg.DemoMode, cfg.JWTSecret)

	if cfg.DemoMode {
		if err := verifier.EnsureDemoAdmin(ctx); err != nil {
			log.Printf("WARNING: ensure demo admin: %v", err)
		} else {
			log.Println("demo admin ready: admin@haloraland.id / admin123")
		}
	}
	auditLog := audit.New(pool, 1024)
	defer auditLog.Close()

	srv := server.New(server.Deps{
		Cfg: cfg, Pool: pool, Verifier: verifier, Audit: auditLog,
		Limiter: limiter, AHSPPath: defaultAHSPPath,
	})

	httpSrv := &http.Server{
		Addr:              cfg.ParsePort(),
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("halora-be listening on %s", cfg.ParsePort())
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutCtx)
}

func migrate(cfg *config.Config) {
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := migration.Run(ctx, pool); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied")
}

func importAHSP(cfg *config.Config) {
	fs := flag.NewFlagSet("import-ahsp", flag.ExitOnError)
	file := fs.String("file", defaultAHSPPath, "path to ahsp xlsx")
	sheet := fs.String("sheet", "", "sheet name (empty = all importable sheets)")
	force := fs.Bool("force", false, "delete existing sheet rows before reimport")
	_ = fs.Parse(os.Args[2:])

	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	im := ahsp.NewImporter(pool)
	f, err := openFile(*file)
	if err != nil {
		log.Fatal("open xlsx: ", err)
	}
	defer f.Close()

	sheets := ahsp.ListSheets(f)
	if *sheet != "" {
		sheets = []string{*sheet}
	}
	for _, sh := range sheets {
		items, err := ahsp.ParseSheet(f, sh)
		if err != nil {
			log.Printf("parse %s: %v", sh, err)
			continue
		}
		res, err := im.ImportSheet(ctx, items, *force)
		if err != nil {
			log.Printf("import %s: %v", sh, err)
			continue
		}
		log.Printf("sheet=%s items=%d rincian=%d skipped=%d", res.Sheet, res.Items, res.Rincian, res.Skipped)
	}
}

func seed(cfg *config.Config) {
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := migration.Run(ctx, pool); err != nil {
		log.Fatal(err)
	}
	if err := seedDB(ctx, pool); err != nil {
		log.Fatal(err)
	}
	log.Println("seed complete")
}
