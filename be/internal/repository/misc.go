package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// RekapRepo manages the per-project margin settings row (kategori='settings').
type RekapRepo struct{ pool *pgxpool.Pool }

func NewRekapRepo(pool *pgxpool.Pool) *RekapRepo { return &RekapRepo{pool: pool} }

func (r *RekapRepo) GetMargin(ctx context.Context, proyekID int32) (decimal.Decimal, error) {
	var m sql.NullString
	err := r.pool.QueryRow(ctx, `SELECT margin::text FROM rekap WHERE "proyekId" = $1 AND kategori = 'settings'`, proyekID).Scan(&m)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return decimal.Zero, err
	}
	if !m.Valid {
		return decimal.Zero, nil
	}
	return scanDec(m.String), nil
}

func (r *RekapRepo) UpsertMargin(ctx context.Context, proyekID int32, margin decimal.Decimal) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rekap ("proyekId", kategori, uraian, urutan, margin)
		VALUES ($1, 'settings', 'Margin & Overhead Settings', 0, $2)
		ON CONFLICT ("proyekId", kategori)
		DO UPDATE SET margin = EXCLUDED.margin, "updatedAt" = CURRENT_TIMESTAMP`,
		proyekID, decArg(margin))
	return err
}

// --- Invoice ---

type InvoiceRepo struct{ pool *pgxpool.Pool }

func NewInvoiceRepo(pool *pgxpool.Pool) *InvoiceRepo { return &InvoiceRepo{pool: pool} }

func (r *InvoiceRepo) List(ctx context.Context, proyekID int32) ([]models.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "proyekId", nomor, tanggal, total::text, status, "createdAt", "updatedAt"
		FROM invoice WHERE "proyekId" = $1 ORDER BY id DESC`, proyekID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Invoice
	for rows.Next() {
		var inv models.Invoice
		var total string
		if err := rows.Scan(&inv.ID, &inv.ProyekID, &inv.Nomor, &inv.Tanggal, &total, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		inv.Total = scanDec(total)
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *InvoiceRepo) Create(ctx context.Context, proyekID int32, tanggal string, total decimal.Decimal, status models.StatusInvoice) (*models.Invoice, error) {
	var seq int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)+1 FROM invoice WHERE "proyekId" = $1`, proyekID).Scan(&seq); err != nil {
		return nil, err
	}
	nomor := "INV-" + strconv.Itoa(int(proyekID)) + "-" + pad4(seq)
	row := r.pool.QueryRow(ctx, `
		INSERT INTO invoice ("proyekId", nomor, tanggal, total, status)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, "proyekId", nomor, tanggal, total::text, status, "createdAt", "updatedAt"`,
		proyekID, nomor, tanggal, decArg(total), status)
	var inv models.Invoice
	var totalStr string
	if err := row.Scan(&inv.ID, &inv.ProyekID, &inv.Nomor, &inv.Tanggal, &totalStr, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
		return nil, err
	}
	inv.Total = scanDec(totalStr)
	return &inv, nil
}

func pad4(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// --- Logistik ---

type LogistikRepo struct{ pool *pgxpool.Pool }

func NewLogistikRepo(pool *pgxpool.Pool) *LogistikRepo { return &LogistikRepo{pool: pool} }

func (r *LogistikRepo) List(ctx context.Context, proyekID int32) ([]models.Logistik, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "proyekId", "namaMaterial", satuan, volume::text, "hargaSatuan"::text, "totalBiaya"::text,
		tanggal, keterangan, "createdAt", "updatedAt"
		FROM logistik WHERE "proyekId" = $1 ORDER BY id DESC`, proyekID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Logistik
	for rows.Next() {
		var l models.Logistik
		var vol, hs, tb string
		var tanggal sql.NullTime
		var ket sql.NullString
		if err := rows.Scan(&l.ID, &l.ProyekID, &l.NamaMaterial, &l.Satuan, &vol, &hs, &tb, &tanggal, &ket, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		l.Volume = scanDec(vol)
		l.HargaSatuan = scanDec(hs)
		l.TotalBiaya = scanDec(tb)
		if tanggal.Valid {
			t := tanggal.Time
			l.Tanggal = &t
		}
		l.Keterangan = strPtr(ket)
		out = append(out, l)
	}
	return out, rows.Err()
}

// --- Realisasi ---

type RealisasiRepo struct{ pool *pgxpool.Pool }

func NewRealisasiRepo(pool *pgxpool.Pool) *RealisasiRepo { return &RealisasiRepo{pool: pool} }

func (r *RealisasiRepo) List(ctx context.Context, proyekID int32) ([]models.Realisasi, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "proyekId", tanggal, kategori, jumlah::text, keterangan, "createdAt", "updatedAt"
		FROM realisasi WHERE "proyekId" = $1 ORDER BY tanggal DESC`, proyekID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Realisasi
	for rows.Next() {
		var re models.Realisasi
		var jum string
		var ket sql.NullString
		if err := rows.Scan(&re.ID, &re.ProyekID, &re.Tanggal, &re.Kategori, &jum, &ket, &re.CreatedAt, &re.UpdatedAt); err != nil {
			return nil, err
		}
		re.Jumlah = scanDec(jum)
		re.Keterangan = strPtr(ket)
		out = append(out, re)
	}
	return out, rows.Err()
}

// --- Feedback + News ---

type FeedbackRepo struct{ pool *pgxpool.Pool }

func NewFeedbackRepo(pool *pgxpool.Pool) *FeedbackRepo { return &FeedbackRepo{pool: pool} }

func (r *FeedbackRepo) ListByUser(ctx context.Context, userID int32) ([]models.Feedback, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, "userId", subject, message, status, "createdAt", "updatedAt" FROM feedback WHERE "userId" = $1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.ID, &f.UserID, &f.Subject, &f.Message, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeedbackRepo) Create(ctx context.Context, userID int32, subject, message string) (*models.Feedback, error) {
	var f models.Feedback
	err := r.pool.QueryRow(ctx, `
		INSERT INTO feedback ("userId", subject, message) VALUES ($1,$2,$3)
		RETURNING id, "userId", subject, message, status, "createdAt", "updatedAt"`, userID, subject, message).
		Scan(&f.ID, &f.UserID, &f.Subject, &f.Message, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

type NewsRepo struct{ pool *pgxpool.Pool }

func NewNewsRepo(pool *pgxpool.Pool) *NewsRepo { return &NewsRepo{pool: pool} }

func (r *NewsRepo) ListActive(ctx context.Context) ([]models.News, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, content, "isActive", "createdAt", "updatedAt" FROM news WHERE "isActive" = true ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.IsActive, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NewsRepo) ListAll(ctx context.Context) ([]models.News, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, content, "isActive", "createdAt", "updatedAt" FROM news ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.IsActive, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NewsRepo) Create(ctx context.Context, title, content string) (*models.News, error) {
	var n models.News
	err := r.pool.QueryRow(ctx, `
		INSERT INTO news (title, content, "isActive") VALUES ($1, $2, true)
		RETURNING id, title, content, "isActive", "createdAt", "updatedAt"`, title, content).
		Scan(&n.ID, &n.Title, &n.Content, &n.IsActive, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepo) Update(ctx context.Context, id int32, title, content *string) (*models.News, error) {
	var n models.News
	err := r.pool.QueryRow(ctx, `
		UPDATE news SET
			title = COALESCE($2, title),
			content = COALESCE($3, content),
			"updatedAt" = NOW()
		WHERE id = $1
		RETURNING id, title, content, "isActive", "createdAt", "updatedAt"`,
		id, title, content).
		Scan(&n.ID, &n.Title, &n.Content, &n.IsActive, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM news WHERE id = $1`, id)
	return err
}
