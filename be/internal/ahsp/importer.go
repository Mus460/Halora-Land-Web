package ahsp

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/models"
)

// Importer batch-inserts parsed AHSP work items into master_analisa +
// rincian_analisa inside one transaction (§3.4, §5.3 porting note).
// Idempotent via ahsp_kode + is_system; forceReimport deletes the sheet first.
type Importer struct{ pool *pgxpool.Pool }

func NewImporter(pool *pgxpool.Pool) *Importer { return &Importer{pool: pool} }

type ImportResult struct {
	Sheet       string
	Items       int
	Rincian     int
	Skipped     int
}

func (im *Importer) ImportSheet(ctx context.Context, items []WorkItem, force bool) (*ImportResult, error) {
	if len(items) == 0 {
		return &ImportResult{}, nil
	}
	sheet := items[0].Sheet

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
		err := tx.QueryRow(ctx, `SELECT id FROM master_analisa WHERE "ahspKode" = $1 AND "isSystem" = true LIMIT 1`, it.Kode).
			Scan(&existingID)
		if err == nil {
			res.Skipped++
			continue
		}

		kategori := it.Section
		satuan := it.Satuan
		var masterID int32
		if err := tx.QueryRow(ctx, `
			INSERT INTO master_analisa (kode, nama, level, satuan, "hargaSatuan", kategori, "isGlobal", "isSystem", "ahspKode", "ahspSheet", "biayaUmum")
			VALUES ($1,$2,4,$3,$4,$5,true,true,$6,$7,$8)
			RETURNING id`,
			it.Kode, it.Nama, satuan, it.HargaSatuan.String(), kategori, it.Kode, sheet, it.BiayaUmum.String()).
			Scan(&masterID); err != nil {
			return nil, fmt.Errorf("insert master_analisa %s: %w", it.Kode, err)
		}
		res.Items++

		for i, b := range it.Breakdown {
			if b.Nama == "" {
				continue
			}
			tipe := models.TipeKomponen(b.Tipe)
			if _, err := tx.Exec(ctx, `
				INSERT INTO rincian_analisa ("masterAnalisaId", "komponenId", koef, tipe, nama, satuan, "hargaSatuan", "jumlahHarga", "kodeReferensi", urutan)
				VALUES ($1,NULL,$2,$3,$4,$5,$6,$7,$8,$9)`,
				masterID, b.Koef.String(), tipe, b.Nama, b.Satuan,
				b.HargaSatuan.String(), b.JumlahHarga.String(), b.KodeReferensi, i); err != nil {
				return nil, fmt.Errorf("insert rincian: %w", err)
			}
			res.Rincian++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
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
