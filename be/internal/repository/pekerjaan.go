package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

type PekerjaanRepo struct{ pool *pgxpool.Pool }

func NewPekerjaanRepo(pool *pgxpool.Pool) *PekerjaanRepo { return &PekerjaanRepo{pool: pool} }

type ListPekerjaanFilter struct {
	ProyekID  *int32
	Kategori  string
	Search    string
}

func (r *PekerjaanRepo) List(ctx context.Context, f ListPekerjaanFilter) ([]models.Pekerjaan, error) {
	q := `SELECT id, "proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
		"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu, (waktu * volume) AS "totalWaktu",
		"createdAt", "updatedAt"
		FROM pekerjaan WHERE 1=1`
	var args []any
	if f.ProyekID != nil {
		args = append(args, *f.ProyekID)
		q += ` AND "proyekId" = $1`
	}
	if f.Kategori != "" {
		args = append(args, f.Kategori)
		q += ` AND kategori = $` + strconv.Itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += ` AND "uraianPekerjaan" ILIKE $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Pekerjaan
	for rows.Next() {
		p, err := scanPekerjaan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func argPlaceholder(i int) string { return strconv.Itoa(i) }

func scanPekerjaan(s rowScanner) (*models.Pekerjaan, error) {
	var p models.Pekerjaan
	var vol, hs, tb string
	var level, tipe, waktu, totalWaktu sql.NullString
	var maID sql.NullInt32
	if err := s.Scan(&p.ID, &p.ProyekID, &p.Kategori, &p.UraianPekerjaan, &vol, &p.Satuan, &hs, &tb,
		&p.MetodeHitung, &level, &tipe, &maID, &waktu, &totalWaktu, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Volume = scanDec(vol)
	p.HargaSatuan = scanDec(hs)
	p.TotalBiaya = scanDec(tb)
	p.LevelPekerjaan = strPtr(level)
	p.TipePekerjaan = strPtr(tipe)
	p.MasterAnalisaID = i32Ptr(maID)
	p.Waktu = scanDecPtr(waktu)
	p.TotalWaktu = scanDecPtr(totalWaktu)
	return &p, nil
}

func (r *PekerjaanRepo) Get(ctx context.Context, id int32) (*models.Pekerjaan, error) {
	p, err := scanPekerjaan(r.pool.QueryRow(ctx, `
		SELECT id, "proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
		"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu, (waktu * volume) AS "totalWaktu",
		"createdAt", "updatedAt"
		FROM pekerjaan WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	details, err := r.ListDetailAnalisa(ctx, id)
	if err != nil {
		return nil, err
	}
	p.DetailAnalisa = details
	return p, nil
}

func (r *PekerjaanRepo) GetByID(ctx context.Context, id int32) (*models.Pekerjaan, error) {
	p, err := scanPekerjaan(r.pool.QueryRow(ctx, `
		SELECT id, "proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
		"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu, (waktu * volume) AS "totalWaktu",
		"createdAt", "updatedAt"
		FROM pekerjaan WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// WaktuKoef returns the time coefficient (jam per satuan) for a master analisa
// item, computed as the sum of waktu over its rincian rows. Nil when absent.
func (r *PekerjaanRepo) WaktuKoef(ctx context.Context, masterAnalisaID int32) (*decimal.Decimal, error) {
	var s sql.NullString
	err := r.pool.QueryRow(ctx, `
		SELECT SUM(waktu) FROM rincian_analisa WHERE "masterAnalisaId" = $1`, masterAnalisaID).Scan(&s)
	if err != nil {
		return nil, err
	}
	return scanDecPtr(s), nil
}

type CreatePekerjaanInput struct {
	ProyekID        int32
	Kategori        models.KategoriPekerjaan
	UraianPekerjaan string
	Volume          decimal.Decimal
	Satuan          string
	HargaSatuan     decimal.Decimal
	TotalBiaya      decimal.Decimal
	MetodeHitung    models.MetodeHitung
	LevelPekerjaan  *string
	TipePekerjaan   *string
	MasterAnalisaID *int32
	Waktu           *decimal.Decimal
}

func (r *PekerjaanRepo) Create(ctx context.Context, tx pgx.Tx, in CreatePekerjaanInput) (*models.Pekerjaan, error) {
	exec := r.execer(tx)
	row := exec.QueryRow(ctx, `
		INSERT INTO pekerjaan ("proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
			"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, "proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
			"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu, (waktu * volume) AS "totalWaktu",
			"createdAt", "updatedAt"`,
		in.ProyekID, in.Kategori, in.UraianPekerjaan, decArg(in.Volume), in.Satuan, decArg(in.HargaSatuan), decArg(in.TotalBiaya),
		in.MetodeHitung, in.LevelPekerjaan, in.TipePekerjaan, in.MasterAnalisaID, decPtrArg(in.Waktu))
	return scanPekerjaan(row)
}

type UpdatePekerjaanInput struct {
	Volume         *decimal.Decimal
	HargaSatuan    *decimal.Decimal
	TotalBiaya     *decimal.Decimal
	Uraian         *string
	Satuan         *string
	LevelPekerjaan *string
	TipePekerjaan  *string
	MetodeHitung   *models.MetodeHitung
}

func (r *PekerjaanRepo) Update(ctx context.Context, id int32, in UpdatePekerjaanInput) (*models.Pekerjaan, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pekerjaan SET
			volume = COALESCE($2, volume),
			"hargaSatuan" = COALESCE($3, "hargaSatuan"),
			"totalBiaya" = COALESCE($4, "totalBiaya"),
			"uraianPekerjaan" = COALESCE($5, "uraianPekerjaan"),
			satuan = COALESCE($6, satuan),
			"levelPekerjaan" = COALESCE($7, "levelPekerjaan"),
			"tipePekerjaan" = COALESCE($8, "tipePekerjaan"),
			"metodeHitung" = COALESCE($9, "metodeHitung"),
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, "proyekId", kategori, "uraianPekerjaan", volume, satuan, "hargaSatuan", "totalBiaya",
			"metodeHitung", "levelPekerjaan", "tipePekerjaan", "masterAnalisaId", waktu, (waktu * volume) AS "totalWaktu",
			"createdAt", "updatedAt"`,
		id, decPtrArg(in.Volume), decPtrArg(in.HargaSatuan), decPtrArg(in.TotalBiaya),
		in.Uraian, in.Satuan, in.LevelPekerjaan, in.TipePekerjaan, in.MetodeHitung)
	return scanPekerjaan(row)
}

func (r *PekerjaanRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM pekerjaan WHERE id = $1`, id)
	return err
}

func (r *PekerjaanRepo) execer(tx pgx.Tx) executor {
	if tx != nil {
		return tx
	}
	return r.pool
}

type executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type CreateDetailInput struct {
	PekerjaanID     int32
	MasterHargaID   *int32
	MasterAnalisaID *int32
	Nama            string
	Satuan          string
	Koef            decimal.Decimal
	HargaSatuan     decimal.Decimal
	TotalBiaya      decimal.Decimal
	Tipe            models.TipeKomponen
	SourceKode      *string
}

func (r *PekerjaanRepo) CreateDetail(ctx context.Context, tx pgx.Tx, in CreateDetailInput) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `
		INSERT INTO detail_analisa ("pekerjaanId", "masterHargaId", "masterAnalisaId", nama, satuan, koef,
			"hargaSatuan", "totalBiaya", tipe, "snapshotAt", "sourceKode")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10)`,
		in.PekerjaanID, in.MasterHargaID, in.MasterAnalisaID, in.Nama, in.Satuan, decArg(in.Koef),
		decArg(in.HargaSatuan), decArg(in.TotalBiaya), in.Tipe, in.SourceKode)
	return err
}

func (r *PekerjaanRepo) ListDetailAnalisa(ctx context.Context, pekerjaanID int32) ([]models.DetailAnalisa, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "pekerjaanId", "masterHargaId", "masterAnalisaId", nama, satuan, koef, "hargaSatuan",
			"totalBiaya", tipe, "snapshotAt", "sourceKode"
		FROM detail_analisa WHERE "pekerjaanId" = $1 ORDER BY id ASC`, pekerjaanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DetailAnalisa
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanDetail(s rowScanner) (*models.DetailAnalisa, error) {
	var d models.DetailAnalisa
	var mhID, maID sql.NullInt32
	var koef, hs, tb string
	var src sql.NullString
	if err := s.Scan(&d.ID, &d.PekerjaanID, &mhID, &maID, &d.Nama, &d.Satuan, &koef, &hs, &tb, &d.Tipe, &d.SnapshotAt, &src); err != nil {
		return nil, err
	}
	d.MasterHargaID = i32Ptr(mhID)
	d.MasterAnalisaID = i32Ptr(maID)
	d.Koef = scanDec(koef)
	d.HargaSatuan = scanDec(hs)
	d.TotalBiaya = scanDec(tb)
	d.SourceKode = strPtr(src)
	return &d, nil
}

// DeleteDetails removes all frozen breakdown rows for a pekerjaan (used by recalculate).
func (r *PekerjaanRepo) DeleteDetails(ctx context.Context, tx pgx.Tx, pekerjaanID int32) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `DELETE FROM detail_analisa WHERE "pekerjaanId" = $1`, pekerjaanID)
	return err
}

func (r *PekerjaanRepo) SetTotal(ctx context.Context, tx pgx.Tx, id int32, hs, tb decimal.Decimal) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `UPDATE pekerjaan SET "hargaSatuan" = $2, "totalBiaya" = $3, "updatedAt" = NOW() WHERE id = $1`,
		id, decArg(hs), decArg(tb))
	return err
}
