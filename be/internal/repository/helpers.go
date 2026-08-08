package repository

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// pgx scans Postgres DECIMAL columns into string, from which we build
// shopspring/decimal values (ARCHITECTURE.md §3.7: never use float64 for money).

func scanDec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func scanDecPtr(s sql.NullString) *decimal.Decimal {
	if !s.Valid || s.String == "" {
		return nil
	}
	d, err := decimal.NewFromString(s.String)
	if err != nil {
		return nil
	}
	return &d
}

func strPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func timePtr(s sql.NullString) *time.Time {
	if !s.Valid {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return nil
	}
	return &t
}

func i32Ptr(s sql.NullInt32) *int32 {
	if !s.Valid {
		return nil
	}
	v := s.Int32
	return &v
}

func decPtrArg(d *decimal.Decimal) any {
	if d == nil {
		return nil
	}
	return d.String()
}

func decArg(d decimal.Decimal) any { return d.String() }
