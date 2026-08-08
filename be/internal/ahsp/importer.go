package ahsp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// Importer batch-inserts parsed AHSP work items into analysis_masters +
// analysis_components inside one transaction. Idempotent via ahsp_kode +
// is_system; forceReimport deletes the sheet first.
type Importer struct{ pool *pgxpool.Pool }

func NewImporter(pool *pgxpool.Pool) *Importer { return &Importer{pool: pool} }

type ImportResult struct {
	Sheet      string
	Items      int
	Components int
	Skipped    int
}

// ImportPriceList upserts the "Upah & Bahan" price list into price_masters
// (global/system rows, deduped by name+category). Returns rows inserted.
func (im *Importer) ImportPriceList(ctx context.Context, prices []PriceItem, force bool) (int, error) {
	if len(prices) == 0 {
		return 0, nil
	}
	tx, err := im.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if force {
		if _, err := tx.Exec(ctx, `DELETE FROM price_masters WHERE "isSystem" = true AND "isGlobal" = true`); err != nil {
			return 0, err
		}
	}

	inserted := 0
	for _, p := range prices {
		var id int32
		err := tx.QueryRow(ctx, `
			SELECT id FROM price_masters
			WHERE name = $1 AND unit = $2 AND type = $3 AND "isGlobal" = true AND "userId" IS NULL
			LIMIT 1`, p.Name, p.Unit, models.ComponentType(p.Type)).Scan(&id)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return inserted, err
		}
		var ahspCode any
		if p.Code != "" {
			ahspCode = p.Code
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO price_masters (name, unit, price, type, "isGlobal", "isSystem", "ahspCode")
			VALUES ($1,$2,$3,$4,true,true,$5)`,
			p.Name, p.Unit, p.Price.String(), models.ComponentType(p.Type), ahspCode); err != nil {
			return inserted, err
		}
		inserted++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return inserted, nil
}

// ImportSheet inserts work items + components breakdowns. Components prices are
// resolved from the price list (by referenceCode, fallback by name); rows
// named "Tukang batu" also get duration = 1/coefficient. Item unitPrice is computed
// as sum(totalPrice) x (1 + generalCost).
func (im *Importer) ImportSheet(ctx context.Context, items []WorkItem, prices []PriceItem, force bool) (*ImportResult, error) {
	if len(items) == 0 {
		return &ImportResult{}, nil
	}
	sheet := items[0].Sheet

	priceByKode := map[string]PriceItem{}
	priceByName := map[string]PriceItem{}
	priceByNameSat := map[string]PriceItem{}
	norm := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	for _, p := range prices {
		if p.Code != "" {
			priceByKode[norm(p.Code)] = p
		}
		priceByName[norm(p.Name)] = p
		priceByNameSat[norm(p.Name)+"|"+norm(p.Unit)] = p
	}

	tx, err := im.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res := &ImportResult{Sheet: sheet}
	if force {
		if _, err := tx.Exec(ctx, `DELETE FROM analysis_masters WHERE "ahspSheet" = $1 AND "isSystem" = true`, sheet); err != nil {
			return nil, err
		}
	}

	for _, it := range items {
		var existingID int32
		err := tx.QueryRow(ctx, `SELECT id FROM analysis_masters WHERE "ahspCode" = $1 AND "ahspSheet" = $2 AND "isSystem" = true LIMIT 1`, it.Code, sheet).
			Scan(&existingID)
		if err == nil {
			res.Skipped++
			continue
		}

		var masterID int32
		if err := tx.QueryRow(ctx, `
			INSERT INTO analysis_masters (code, name, level, unit, "unitPrice", category, "isGlobal", "isSystem", "ahspCode", "ahspSheet", "generalCost")
			VALUES ($1,$2,4,$3,0,$4,true,true,$5,$6,$7)
			RETURNING id`,
			it.Code, it.Name, it.Unit, it.Section, it.Code, sheet, it.GeneralCost.String()).
			Scan(&masterID); err != nil {
			return nil, fmt.Errorf("insert analysis_masters %s: %w", it.Code, err)
		}
		res.Items++

		// duration = 1/koefisien untuk baris tenaga kerja acuan: "Tukang batu"
		// bila ada, selain itu baris upah dengan coefficient terbesar.
		durationIdx := -1
		for i := range it.Breakdown {
			b := &it.Breakdown[i]
			if b.Type != string(models.ComponentLabor) || b.Coefficient.IsZero() {
				continue
			}
			if durationIdx < 0 {
				durationIdx = i
				continue
			}
			cur := &it.Breakdown[durationIdx]
			isCurBatu := isMason(cur.Name)
			isNewBatu := isMason(b.Name)
			if isNewBatu && !isCurBatu {
				durationIdx = i
			} else if isNewBatu == isCurBatu && b.Coefficient.GreaterThan(cur.Coefficient) {
				durationIdx = i
			}
		}

		total := decimal.Zero
		for i, b := range it.Breakdown {
			if b.Name == "" {
				continue
			}
			p, ok := priceByKode[norm(b.ReferenceCode)]
			if !ok {
				p, ok = priceByNameSat[norm(b.Name)+"|"+norm(b.Unit)]
			}
			if !ok {
				p, ok = priceByName[norm(b.Name)]
			}
			hs := decimal.Zero
			if ok {
				hs = p.Price
			}
			jh := b.Coefficient.Mul(hs).Round(2)
			total = total.Add(jh)

			var duration any
			if i == durationIdx {
				duration = decimal.NewFromInt(1).Div(b.Coefficient).Round(4).String()
			}
			compType := models.ComponentType(b.Type)
			if _, err := tx.Exec(ctx, `
				INSERT INTO analysis_components ("analysisMasterId", "componentId", coefficient, type, name, unit, "unitPrice", "totalPrice", "referenceCode", sequence, duration)
				VALUES ($1,NULL,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
				masterID, b.Coefficient.String(), compType, b.Name, b.Unit,
				hs.String(), jh.String(), b.ReferenceCode, i, duration); err != nil {
				return nil, fmt.Errorf("insert components: %w", err)
			}
			res.Components++
		}

		itemHS := total.Mul(decimal.NewFromInt(1).Add(it.GeneralCost)).Round(2)
		if _, err := tx.Exec(ctx, `UPDATE analysis_masters SET "unitPrice" = $1 WHERE id = $2`, itemHS.String(), masterID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func isMason(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "Tukang batu")
}

// ImportStatus reports per-sheet whether it has been imported and the row count.
func (im *Importer) ImportStatus(ctx context.Context) (map[string]ImportStatus, error) {
	rows, err := im.pool.Query(ctx, `SELECT "ahspSheet", count(*) FROM analysis_masters WHERE "isSystem" = true GROUP BY "ahspSheet"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]ImportStatus{}
	for rows.Next() {
		var sheet string
		var n int
		if err := rows.Scan(&sheet, &n); err != nil {
			return nil, err
		}
		out[sheet] = ImportStatus{Sheet: sheet, Imported: true, Count: n}
	}
	return out, rows.Err()
}

type ImportStatus struct {
	Sheet    string `json:"sheet"`
	Imported bool   `json:"imported"`
	Count    int    `json:"count"`
}
