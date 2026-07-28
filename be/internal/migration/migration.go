package migration

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var fs embed.FS

// Run applies all embedded SQL migrations in lexical order. Idempotency is
// delegated to the SQL itself (CREATE EXTENSION IF NOT EXISTS, etc.). For a
// greenfield DB this is a single schema file; the runner is forward-compatible
// with a golang-migrate-style sequence of numbered files.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		b, err := fs.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}
