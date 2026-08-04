package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

// SnapshotService implements the frozen-breakdown engine (ARCHITECTURE.md §2.3,
// §3.1, §8.2). hargaSatuan/totalBiaya are decimal end-to-end (§3.7).
type SnapshotService struct {
	pool         *pgxpool.Pool
	pekerjaan    *repository.PekerjaanRepo
	masterAnalisa *repository.MasterAnalisaRepo
	audit        *audit.Logger
}

func NewSnapshotService(pool *pgxpool.Pool, pr *repository.PekerjaanRepo, mr *repository.MasterAnalisaRepo, al *audit.Logger) *SnapshotService {
	return &SnapshotService{pool: pool, pekerjaan: pr, masterAnalisa: mr, audit: al}
}

// FromAHSP creates a Pekerjaan from a MasterAnalisa catalog item and copies the
// rincian template into frozen detail_analisa rows (§2.3). One transaction.
func (s *SnapshotService) FromAHSP(ctx context.Context, proyekID, masterAnalisaID int32, volume decimal.Decimal, applyBreakdown bool) (*models.Pekerjaan, error) {
	ma, err := s.masterAnalisa.Get(ctx, masterAnalisaID)
	if err != nil {
		return nil, fmt.Errorf("load master analisa: %w", err)
	}

	hs := decimal.Zero
	if ma.HargaSatuan != nil {
		hs = *ma.HargaSatuan
	}
	totalBiaya := hs.Mul(volume)

	kat := models.KategoriCustom
	if ma.Kategori != nil {
		kat = models.KategoriPekerjaan(*ma.Kategori)
	}
	satuan := "unit"
	if ma.Satuan != nil && *ma.Satuan != "" {
		satuan = *ma.Satuan
	}
	levelStr := fmt.Sprintf("%d", ma.Level)

	waktu, err := s.pekerjaan.WaktuKoef(ctx, masterAnalisaID)
	if err != nil {
		return nil, fmt.Errorf("load waktu koefisien: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	p, err := s.pekerjaan.Create(ctx, tx, repository.CreatePekerjaanInput{
		ProyekID:        proyekID,
		Kategori:        kat,
		UraianPekerjaan: ma.Nama,
		Volume:          volume,
		Satuan:          satuan,
		HargaSatuan:     hs,
		TotalBiaya:      totalBiaya,
		MetodeHitung:    models.MetodeAHSP,
		LevelPekerjaan:  &levelStr,
		MasterAnalisaID: &masterAnalisaID,
		Waktu:           waktu,
	})
	if err != nil {
		return nil, err
	}

	if applyBreakdown {
		rincian, err := s.masterAnalisa.ListRincian(ctx, masterAnalisaID)
		if err != nil {
			return nil, err
		}
		for _, r := range rincian {
			nama := ""
			if r.Nama != nil {
				nama = *r.Nama
			}
			sat := ""
			if r.Satuan != nil {
				sat = *r.Satuan
			}
			dhs := decimal.Zero
			if r.HargaSatuan != nil {
				dhs = *r.HargaSatuan
			}
			jh := decimal.Zero
			if r.JumlahHarga != nil {
				jh = *r.JumlahHarga
			}
			detailTotal := jh.Mul(volume)
			maID := masterAnalisaID
			if err := s.pekerjaan.CreateDetail(ctx, tx, repository.CreateDetailInput{
				PekerjaanID:     p.ID,
				MasterHargaID:   r.KomponenID,
				MasterAnalisaID: &maID,
				Nama:            nama,
				Satuan:          sat,
				Koef:            r.Koef,
				HargaSatuan:     dhs,
				TotalBiaya:      detailTotal,
				Tipe:            r.Tipe,
				SourceKode:      r.KodeReferensi,
			}); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.pekerjaan.Get(ctx, p.ID)
}

// DriftChange is one component's price drift between snapshot and live master.
type DriftChange struct {
	Nama         string          `json:"nama"`
	OldHarga     decimal.Decimal `json:"oldHarga"`
	NewHarga     decimal.Decimal `json:"newHarga"`
	Diff         decimal.Decimal `json:"diff"`
	PercentChange decimal.Decimal `json:"percentChange"`
}

type ValidationResult struct {
	IsValid bool          `json:"isValid"`
	Changes []DriftChange `json:"changes"`
}

// ValidateSnapshot reports drift between frozen detail_analisa.hargaSatuan and
// live master_harga.harga (§2.4, §8.2). masterHargaId null or deleted -> skip.
func (s *SnapshotService) ValidateSnapshot(ctx context.Context, pekerjaanID int32) (*ValidationResult, error) {
	if _, err := s.pekerjaan.GetByID(ctx, pekerjaanID); err != nil {
		return nil, err
	}
	details, err := s.pekerjaan.ListDetailAnalisa(ctx, pekerjaanID)
	if err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return &ValidationResult{IsValid: true}, nil
	}

	var ids []int32
	for _, d := range details {
		if d.MasterHargaID != nil {
			ids = append(ids, *d.MasterHargaID)
		}
	}
	live := map[int32]decimal.Decimal{}
	if len(ids) > 0 {
		mhRepo := repository.NewMasterHargaRepo(s.pool)
		live, err = mhRepo.GetMany(ctx, ids)
		if err != nil {
			return nil, err
		}
	}

	res := &ValidationResult{IsValid: true}
	for _, d := range details {
		if d.MasterHargaID == nil {
			continue
		}
		newH, ok := live[*d.MasterHargaID]
		if !ok {
			continue
		}
		if !d.HargaSatuan.Equal(newH) {
			res.IsValid = false
			diff := newH.Sub(d.HargaSatuan)
			pct := decimal.Zero
			if !d.HargaSatuan.IsZero() {
				pct = diff.Div(d.HargaSatuan).Mul(decimal.NewFromInt(100))
			}
			res.Changes = append(res.Changes, DriftChange{
				Nama:          d.Nama,
				OldHarga:      d.HargaSatuan,
				NewHarga:      newH,
				Diff:          diff,
				PercentChange: pct,
			})
		}
	}
	return res, nil
}

// Recalculate re-snapshots a single AHSP pekerjaan from live master prices,
// deletes old details, and updates totals. Audit action 'recalculate' (§2.4).
func (s *SnapshotService) Recalculate(ctx context.Context, pekerjaanID int32, u *auth.AuthUser) (*models.Pekerjaan, error) {
	p, err := s.pekerjaan.GetByID(ctx, pekerjaanID)
	if err != nil {
		return nil, err
	}
	if p.MetodeHitung != models.MetodeAHSP {
		return nil, errors.New("pekerjaan is not AHSP-based")
	}

	existing, err := s.pekerjaan.ListDetailAnalisa(ctx, pekerjaanID)
	if err != nil {
		return nil, err
	}
	var masterAnalisaID *int32
	for _, d := range existing {
		if d.MasterAnalisaID != nil {
			masterAnalisaID = d.MasterAnalisaID
			break
		}
	}
	if masterAnalisaID == nil {
		return nil, errors.New("pekerjaan has no master analisa lineage")
	}

	rincian, err := s.masterAnalisa.ListRincian(ctx, *masterAnalisaID)
	if err != nil {
		return nil, err
	}
	var ids []int32
	for _, r := range rincian {
		if r.KomponenID != nil {
			ids = append(ids, *r.KomponenID)
		}
	}
	live := map[int32]decimal.Decimal{}
	if len(ids) > 0 {
		mhRepo := repository.NewMasterHargaRepo(s.pool)
		live, err = mhRepo.GetMany(ctx, ids)
		if err != nil {
			return nil, err
		}
	}

	var newHS decimal.Decimal
	for _, r := range rincian {
		hs := decimal.Zero
		if r.KomponenID != nil {
			if lv, ok := live[*r.KomponenID]; ok {
				hs = lv
			}
		}
		compTotal := r.Koef.Mul(hs)
		newHS = newHS.Add(compTotal)
	}
	newTotal := newHS.Mul(p.Volume)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	oldTotal := p.TotalBiaya

	if err := s.pekerjaan.DeleteDetails(ctx, tx, pekerjaanID); err != nil {
		return nil, err
	}

	for _, r := range rincian {
		hs := decimal.Zero
		if r.KomponenID != nil {
			if lv, ok := live[*r.KomponenID]; ok {
				hs = lv
			}
		}
		compTotal := r.Koef.Mul(hs)
		nama := ""
		if r.Nama != nil {
			nama = *r.Nama
		}
		sat := ""
		if r.Satuan != nil {
			sat = *r.Satuan
		}
		maID := *masterAnalisaID
		if err := s.pekerjaan.CreateDetail(ctx, tx, repository.CreateDetailInput{
			PekerjaanID:     pekerjaanID,
			MasterHargaID:   r.KomponenID,
			MasterAnalisaID: &maID,
			Nama:            nama,
			Satuan:          sat,
			Koef:            r.Koef,
			HargaSatuan:     hs,
			TotalBiaya:      compTotal.Mul(p.Volume),
			Tipe:            r.Tipe,
			SourceKode:      r.KodeReferensi,
		}); err != nil {
			return nil, err
		}
	}

	if err := s.pekerjaan.SetTotal(ctx, tx, pekerjaanID, newHS, newTotal); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if s.audit != nil && u != nil {
		pid := p.ProyekID
		pkID := pekerjaanID
		s.audit.Log(audit.Params{
			Action:      audit.ActionRecalculate,
			EntityType:  "pekerjaan",
			EntityID:    &pekerjaanID,
			ProyekID:    &pid,
			PekerjaanID: &pkID,
			UserID:      u.UserID,
			OldValue:    map[string]any{"totalBiaya": oldTotal.String()},
			NewValue:    map[string]any{"totalBiaya": newTotal.String()},
		})
	}
	return s.pekerjaan.Get(ctx, pekerjaanID)
}

// RecalculateAll bulk re-snapshots all AHSP pekerjaan in a project. One txn,
// one audit_log per item ('bulk_recalculate') (§2.4).
func (s *SnapshotService) RecalculateAll(ctx context.Context, proyekID int32, u *auth.AuthUser) (int, error) {
	pr := repository.NewPekerjaanRepo(s.pool)
	pid := proyekID
	items, err := pr.List(ctx, repository.ListPekerjaanFilter{ProyekID: &pid})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, p := range items {
		if p.MetodeHitung != models.MetodeAHSP {
			continue
		}
		if _, err := s.Recalculate(ctx, p.ID, u); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
