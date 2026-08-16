package env

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

func TestGetEnvString(t *testing.T) {
	t.Setenv("TEST_STR", "hello")
	if got := GetEnvString("TEST_STR", "dflt"); got != "hello" {
		t.Errorf("got %q want hello", got)
	}
	if got := GetEnvString("TEST_STR_UNSET", "dflt"); got != "dflt" {
		t.Errorf("got %q want dflt", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if got := GetEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("got %d want 42", got)
	}
	if got := GetEnvInt("TEST_INT_UNSET", 7); got != 7 {
		t.Errorf("got %d want 7", got)
	}
	t.Setenv("TEST_INT_BAD", "notanumber")
	if got := GetEnvInt("TEST_INT_BAD", 9); got != 9 {
		t.Errorf("got %d want default 9 for invalid value", got)
	}
}

func TestGetEnvBool(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	if got := GetEnvBool("TEST_BOOL", false); !got {
		t.Error("want true")
	}
	t.Setenv("TEST_BOOL_F", "0")
	if got := GetEnvBool("TEST_BOOL_F", true); got {
		t.Error("want false for '0'")
	}
	if got := GetEnvBool("TEST_BOOL_UNSET", true); !got {
		t.Error("want default true")
	}
	t.Setenv("TEST_BOOL_BAD", "xyz")
	if got := GetEnvBool("TEST_BOOL_BAD", false); got {
		t.Error("want default false for invalid value")
	}
}

func TestGetEnvDecimal(t *testing.T) {
	t.Setenv("TEST_DEC", "0.11")
	got := GetEnvDecimal("TEST_DEC", decimal.Zero)
	if !got.Equal(decimal.RequireFromString("0.11")) {
		t.Errorf("got %s want 0.11", got)
	}
	def := decimal.RequireFromString("0.10")
	if !GetEnvDecimal("TEST_DEC_UNSET", def).Equal(def) {
		t.Error("want default when unset")
	}
	t.Setenv("TEST_DEC_BAD", "abc")
	if !GetEnvDecimal("TEST_DEC_BAD", def).Equal(def) {
		t.Error("want default for invalid decimal")
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost/z")
	t.Setenv("ALLOWED_ORIGINS", "http://a.com,http://b.com")
	t.Setenv("OVERHEAD_RATE", "0.05")
	t.Setenv("PPN_RATE", "0.12")
	t.Setenv("NODE_ENV", "production")
	t.Setenv("JWT_SECRET", "sekret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q want 9090", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://x:y@localhost/z" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "http://b.com" {
		t.Errorf("AllowedOrigins = %v", cfg.AllowedOrigins)
	}
	if !cfg.PPNRate.Equal(decimal.RequireFromString("0.12")) {
		t.Errorf("PPNRate = %s", cfg.PPNRate)
	}
	if !cfg.IsProd {
		t.Error("IsProd = false, want true")
	}
	if cfg.JWTSecret != "sekret" {
		t.Errorf("JWTSecret = %q", cfg.JWTSecret)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://default")
	os.Unsetenv("PORT")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("PPN_RATE")
	os.Unsetenv("NODE_ENV")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q want 8080", cfg.Port)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("AllowedOrigins = %v", cfg.AllowedOrigins)
	}
	if !cfg.PPNRate.Equal(decimal.RequireFromString("0.11")) {
		t.Errorf("PPNRate = %s", cfg.PPNRate)
	}
	if cfg.IsProd {
		t.Error("IsProd = true, want false in development")
	}
}

func TestParsePort(t *testing.T) {
	cases := []struct {
		port string
		want string
	}{
		{"8080", ":8080"},
		{"8080x", ":8080x"}, // non-numeric passthrough
		{"", ":"},
	}
	for _, c := range cases {
		cfg := &Config{Port: c.port}
		if got := cfg.ParsePort(); got != c.want {
			t.Errorf("ParsePort(%q) = %q want %q", c.port, got, c.want)
		}
	}
}

func TestMustEnvPanics(t *testing.T) {
	os.Unsetenv("TEST_REQUIRED")
	defer func() {
		if r := recover(); r == nil {
			t.Error("mustEnv should panic when unset")
		}
	}()
	mustEnv("TEST_REQUIRED")
}
