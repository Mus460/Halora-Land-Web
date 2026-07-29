package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Config holds all application configuration loaded from the environment.
type Config struct {
	Port           string
	DatabaseURL    string
	AllowedOrigins []string

	// Supabase Auth (JWT verification via JWKS)
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseServiceKey string
	SupabaseJWKSURL   string
	SupabaseProjectRef string
	CookieName        string

	// Rollup defaults (configurable per ARCHITECTURE.md §3.5/§8.3 — no longer hardcoded 0.10/0.11)
	OverheadRate decimal.Decimal
	PPNRate      decimal.Decimal

	// DemoMode skips JWT verification and injects a fixed admin user so the
	// app can be demoed without Supabase auth. Enable via DEMO_MODE=true.
	DemoMode  bool
	IsProd    bool

	// JWTSecret signs session tokens in demo (local auth) mode.
	JWTSecret string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:            envOr("PORT", "8080"),
		DatabaseURL:     mustEnv("DATABASE_URL"),
		AllowedOrigins:  strings.Split(envOr("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		SupabaseURL:     mustEnv("NEXT_PUBLIC_SUPABASE_URL"),
		SupabaseAnonKey: mustEnv("NEXT_PUBLIC_SUPABASE_ANON_KEY"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		OverheadRate:    decEnvOr("OVERHEAD_RATE", "0.10"),
		PPNRate:         decEnvOr("PPN_RATE", "0.11"),
		DemoMode:        envOr("DEMO_MODE", "") == "true",
		IsProd:          envOr("NODE_ENV", "development") == "production",
		JWTSecret:       envOr("JWT_SECRET", ""),
	}

	ref := supabaseProjectRef(cfg.SupabaseURL)
	cfg.SupabaseProjectRef = ref
	cfg.SupabaseJWKSURL = fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", strings.TrimRight(cfg.SupabaseURL, "/"))
	cfg.CookieName = fmt.Sprintf("sb-%s-auth-token", ref)

	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func decEnvOr(key, def string) decimal.Decimal {
	return decimal.RequireFromString(envOr(key, def))
}

// supabaseProjectRef extracts the project ref from a Supabase URL like
// https://abcdefgh.supabase.co -> abcdefgh.
func supabaseProjectRef(url string) string {
	parts := strings.SplitN(url, "//", 2)
	if len(parts) != 2 {
		return ""
	}
	host := strings.SplitN(parts[1], ".", 2)
	if len(host) < 1 {
		return ""
	}
	return host[0]
}

// ParsePort returns the numeric port (without leading colon) for net.Listen.
func (c *Config) ParsePort() string {
	if p, err := strconv.Atoi(c.Port); err == nil {
		return fmt.Sprintf(":%d", p)
	}
	return ":" + c.Port
}
