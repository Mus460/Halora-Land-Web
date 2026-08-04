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

// Importer batch-inserts parsed AHSP work items into master_analisa +
// rincian_analisa inside one transaction. Idempotent via ahsp_kode +
// is_system; forceReimport deletes the sheet first.
type Importer struct{ pool *pgxpool.Pool }

func NewImporter(pool *pgxpool.Pool) *Importer { return &Importer{pool: pool} }

type ImportResult struct {
	Sheet   string
	Items   int
	Rincian int
	Skipped int
}

// ImportPriceList upserts the "Upah & Bahan" price list into master_harga
// (global/system rows, deduped by nama+kategori). Returns rows inserted.
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
		if _, err := tx.Exec(ctx, `DELETE FROM master_harga WHERE "isSystem" = true AND "isGlobal" = true`); err != nil {
			return 0, err
		}
	}

	inserted := 0
	for _, p := range prices {
		var id int32
		err := tx.QueryRow(ctx, `
			SELECT id FROM master_harga
			WHERE nama = $1 AND satuan = $2 AND kategori = $3 AND "isGlobal" = true AND "userId" IS NULL
			LIMIT 1`, p.Nama, p.Satuan, models.TipeKomponen(p.Tipe)).Scan(&id)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return inserted, err
		}
		var kodeAHSP any
		if p.Kode != "" {
			kodeAHSP = p.Kode
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO master_harga (nama, satuan, harga, kategori, "isGlobal", "isSystem", "kodeAHSP")
			VALUES ($1,$2,$3,$4,true,true,$5)`,
			p.Nama, p.Satuan, p.Harga.String(), models.TipeKomponen(p.Tipe), kodeAHSP); err != nil {
			return inserted, err
		}
		inserted++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return inserted, nil
}

// ImportSheet inserts work items + rincian breakdowns. Rincian prices are
// resolved from the price list (by kodeReferensi, fallback by name); rows
// named "Tukang batu" also get waktu = 1/koef. Item hargaSatuan is computed
// as sum(jumlahHarga) x (1 + biayaUmum).
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
		if p.Kode != "" {
			priceByKode[norm(p.Kode)] = p
		}
		priceByName[norm(p.Nama)] = p
		priceByNameSat[norm(p.Nama)+"|"+norm(p.Satuan)] = p
	}

	tx, err := im.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res := &ImportResult{Sheet: sheet}
	if force {
		if _, err := tx.Exec(ctx, `DELETE FROM master_analisa WHERE "ahspSheet" = $1 AND "isSystem" = true`, sheet); err != nil {
			return nil, err
		}
	}

	for _, it := range items {
		var existingID int32
		err := tx.QueryRow(ctx, `SELECT id FROM master_analisa WHERE "ahspKode" = $1 AND "ahspSheet" = $2 AND "isSystem" = true LIMIT 1`, it.Kode, sheet).
			Scan(&existingID)
		if err == nil {
			res.Skipped++
			continue
		}

		var masterID int32
		if err := tx.QueryRow(ctx, `
			INSERT INTO master_analisa (kode, nama, level, satuan, "hargaSatuan", kategori, "isGlobal", "isSystem", "ahspKode", "ahspSheet", "biayaUmum")
			VALUES ($1,$2,4,$3,0,$4,true,true,$5,$6,$7)
			RETURNING id`,
			it.Kode, it.Nama, it.Satuan, it.Section, it.Kode, sheet, it.BiayaUmum.String()).
			Scan(&masterID); err != nil {
			return nil, fmt.Errorf("insert master_analisa %s: %w", it.Kode, err)
		}
		res.Items++

		// waktu = 1/koefisien untuk baris tenaga kerja acuan: "Tukang batu"
		// bila ada, selain itu baris upah dengan koef terbesar.
		waktuIdx := -1
		for i := range it.Breakdown {
			b := &it.Breakdown[i]
			if b.Tipe != string(models.KomponenUpah) || b.Koef.IsZero() {
				continue
			}
			if waktuIdx < 0 {
				waktuIdx = i
				continue
			}
			cur := &it.Breakdown[waktuIdx]
			isCurBatu := isTukangBatu(cur.Nama)
			isNewBatu := isTukangBatu(b.Nama)
			if isNewBatu && !isCurBatu {
				waktuIdx = i
			} else if isNewBatu == isCurBatu && b.Koef.GreaterThan(cur.Koef) {
				waktuIdx = i
			}
		}

		total := decimal.Zero
		for i, b := range it.Breakdown {
			if b.Nama == "" {
				continue
			}
			p, ok := priceByKode[norm(b.KodeReferensi)]
			if !ok {
				p, ok = priceByNameSat[norm(b.Nama)+"|"+norm(b.Satuan)]
			}
			if !ok {
				p, ok = priceByName[norm(b.Nama)]
			}
			hs := decimal.Zero
			if ok {
				hs = p.Harga
			}
			jh := b.Koef.Mul(hs).Round(2)
			total = total.Add(jh)

			var waktu any
			if i == waktuIdx {
				waktu = decimal.NewFromInt(1).Div(b.Koef).Round(4).String()
			}
			tipe := models.TipeKomponen(b.Tipe)
			if _, err := tx.Exec(ctx, `
				INSERT INTO rincian_analisa ("masterAnalisaId", "komponenId", koef, tipe, nama, satuan, "hargaSatuan", "jumlahHarga", "kodeReferensi", urutan, waktu)
				VALUES ($1,NULL,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
				masterID, b.Koef.String(), tipe, b.Nama, b.Satuan,
				hs.String(), jh.String(), b.KodeReferensi, i, waktu); err != nil {
				return nil, fmt.Errorf("insert rincian: %w", err)
			}
			res.Rincian++
		}

		itemHS := total.Mul(decimal.NewFromInt(1).Add(it.BiayaUmum)).Round(2)
		if _, err := tx.Exec(ctx, `UPDATE master_analisa SET "hargaSatuan" = $1 WHERE id = $2`, itemHS.String(), masterID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func isTukangBatu(nama string) bool {
	return strings.EqualFold(strings.TrimSpace(nama), "Tukang batu")
}

// ImportStatus reports per-sheet whether it has been imported and the row count.
func (im *Importer) ImportStatus(ctx context.Context) (map[string]ImportStatus, error) {
	rows, err := im.pool.Query(ctx, `SELECT "ahspSheet", count(*) FROM master_analisa WHERE "isSystem" = true GROUP BY "ahspSheet"`)
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
