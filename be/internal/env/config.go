package env

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Config holds all application configuration loaded from the environment.
type Config struct {
	Port           string
	DatabaseURL    string
	AllowedOrigins []string

	// PPN rate applied to the RAB subtotal (configurable per ARCHITECTURE.md §3.5/§8.3).
	PPNRate decimal.Decimal

	IsProd bool

	// JWTSecret signs session tokens issued at login (HS256). Required when
	// serving; auth endpoints reject empty secrets.
	JWTSecret string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:           GetEnvString("PORT", "8080"),
		DatabaseURL:    mustEnv("DATABASE_URL"),
		AllowedOrigins: strings.Split(GetEnvString("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		PPNRate:        GetEnvDecimal("PPN_RATE", decimal.RequireFromString("0.11")),
		IsProd:         GetEnvString("NODE_ENV", "development") == "production",
		JWTSecret:      GetEnvString("JWT_SECRET", ""),
	}

	return cfg, nil
}

// ParsePort returns the numeric port (without leading colon) for net.Listen.
func (c *Config) ParsePort() string {
	if p, err := strconv.Atoi(c.Port); err == nil {
		return fmt.Sprintf(":%d", p)
	}
	return ":" + c.Port
}
