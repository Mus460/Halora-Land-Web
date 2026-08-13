package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

// SnapshotService implements the frozen-breakdown engine (ARCHITECTURE.md §2.3,
// §3.1, §8.2). unitPrice/totalCost are decimal end-to-end (§3.7).
type SnapshotService struct {
	pool          database.Pool
	work_items    *repository.WorkItemRepo
	masterAnalisa *repository.AnalysisMasterRepo
	audit         *audit.Logger
}

func NewSnapshotService(pool database.Pool, pr *repository.WorkItemRepo, mr *repository.AnalysisMasterRepo, al *audit.Logger) *SnapshotService {
	return &SnapshotService{pool: pool, work_items: pr, masterAnalisa: mr, audit: al}
}

// FromAHSP creates a WorkItem from a AnalysisMaster catalog item and copies the
// components template into frozen work_item_details rows (§2.3). One transaction.
// A non-nil category overrides the master's category (used by the per-category
// work item pages); nil falls back to the master's own category.
func (s *SnapshotService) FromAHSP(ctx context.Context, projectID, analysisMasterID int32, volume decimal.Decimal, applyBreakdown bool, category *models.WorkCategory) (*models.WorkItem, error) {
	ma, err := s.masterAnalisa.Get(ctx, analysisMasterID)
	if err != nil {
		return nil, fmt.Errorf("load master analisa: %w", err)
	}

	hs := decimal.Zero
	if ma.UnitPrice != nil {
		hs = *ma.UnitPrice
	}
	totalCost := hs.Mul(volume)

	kat := models.CategoryCustom
	if category != nil {
		kat = *category
	} else if ma.Category != nil {
		kat = models.WorkCategory(*ma.Category)
	}
	unit := "unit"
	if ma.Unit != nil && *ma.Unit != "" {
		unit = *ma.Unit
	}
	levelStr := fmt.Sprintf("%d", ma.Level)

	duration, err := s.work_items.DurationCoefficient(ctx, analysisMasterID)
	if err != nil {
		return nil, fmt.Errorf("load duration coefficient: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	p, err := s.work_items.Create(ctx, tx, repository.CreateWorkItemInput{
		ProjectID:         projectID,
		Category:          kat,
		Description:       ma.Name,
		Volume:            volume,
		Unit:              unit,
		UnitPrice:         hs,
		TotalCost:         totalCost,
		CalculationMethod: models.MethodAHSP,
		Level:             &levelStr,
		AnalysisMasterID:  &analysisMasterID,
		BasePrice:         ma.UnitPrice,
		Duration:          duration,
	})
	if err != nil {
		return nil, err
	}

	if applyBreakdown {
		if err := s.applyBreakdown(ctx, tx, p.ID, analysisMasterID, volume); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.work_items.Get(ctx, p.ID)
}

// applyBreakdown copies the master's components into frozen work_item_details
// rows for a work item (used by FromAHSP and Create).
func (s *SnapshotService) applyBreakdown(ctx context.Context, tx pgx.Tx, workItemID, analysisMasterID int32, volume decimal.Decimal) error {
	components, err := s.masterAnalisa.ListComponents(ctx, analysisMasterID)
	if err != nil {
		return err
	}
	for _, r := range components {
		name := ""
		if r.Name != nil {
			name = *r.Name
		}
		sat := ""
		if r.Unit != nil {
			sat = *r.Unit
		}
		dhs := decimal.Zero
		if r.UnitPrice != nil {
			dhs = *r.UnitPrice
		}
		jh := decimal.Zero
		if r.TotalPrice != nil {
			jh = *r.TotalPrice
		}
		detailTotal := jh.Mul(volume)
		maID := analysisMasterID
		if err := s.work_items.CreateDetail(ctx, tx, repository.CreateDetailInput{
			WorkItemID:       workItemID,
			PriceMasterID:    r.ComponentID,
			AnalysisMasterID: &maID,
			Name:             name,
			Unit:             sat,
			Coefficient:      r.Coefficient,
			UnitPrice:        dhs,
			TotalCost:        detailTotal,
			Type:             r.Type,
			SourceCode:       r.ReferenceCode,
		}); err != nil {
			return err
		}
	}
	return nil
}

// DriftChange is one component's price drift between snapshot and live master.
type DriftChange struct {
	Name          string          `json:"name"`
	OldPrice      decimal.Decimal `json:"oldPrice"`
	NewPrice      decimal.Decimal `json:"newPrice"`
	Diff          decimal.Decimal `json:"diff"`
	PercentChange decimal.Decimal `json:"percentChange"`
}

type ValidationResult struct {
	IsValid bool          `json:"isValid"`
	Changes []DriftChange `json:"changes"`
}

// ValidateSnapshot reports drift between frozen work_item_details.unitPrice and
// live price_masters.price (§2.4, §8.2). priceMasterId null or deleted -> skip.
func (s *SnapshotService) ValidateSnapshot(ctx context.Context, workItemID int32) (*ValidationResult, error) {
	if _, err := s.work_items.GetByID(ctx, workItemID); err != nil {
		return nil, err
	}
	details, err := s.work_items.ListWorkItemDetail(ctx, workItemID)
	if err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return &ValidationResult{IsValid: true}, nil
	}

	var ids []int32
	for _, d := range details {
		if d.PriceMasterID != nil {
			ids = append(ids, *d.PriceMasterID)
		}
	}
	live := map[int32]decimal.Decimal{}
	if len(ids) > 0 {
		mhRepo := repository.NewPriceMasterRepo(s.pool)
		live, err = mhRepo.GetMany(ctx, ids)
		if err != nil {
			return nil, err
		}
	}

	res := &ValidationResult{IsValid: true}
	for _, d := range details {
		if d.PriceMasterID == nil {
			continue
		}
		newH, ok := live[*d.PriceMasterID]
		if !ok {
			continue
		}
		if !d.UnitPrice.Equal(newH) {
			res.IsValid = false
			diff := newH.Sub(d.UnitPrice)
			pct := decimal.Zero
			if !d.UnitPrice.IsZero() {
				pct = diff.Div(d.UnitPrice).Mul(decimal.NewFromInt(100))
			}
			res.Changes = append(res.Changes, DriftChange{
				Name:          d.Name,
				OldPrice:      d.UnitPrice,
				NewPrice:      newH,
				Diff:          diff,
				PercentChange: pct,
			})
		}
	}
	return res, nil
}

// Recalculate re-snapshots a single AHSP work_items from live master prices,
// deletes old details, and updates totals. Audit action 'recalculate' (§2.4).
func (s *SnapshotService) Recalculate(ctx context.Context, workItemID int32, u *auth.AuthUser) (*models.WorkItem, error) {
	p, err := s.work_items.GetByID(ctx, workItemID)
	if err != nil {
		return nil, err
	}
	if p.CalculationMethod != models.MethodAHSP {
		return nil, errors.New("work_items is not AHSP-based")
	}

	existing, err := s.work_items.ListWorkItemDetail(ctx, workItemID)
	if err != nil {
		return nil, err
	}
	var analysisMasterID *int32
	for _, d := range existing {
		if d.AnalysisMasterID != nil {
			analysisMasterID = d.AnalysisMasterID
			break
		}
	}
	if analysisMasterID == nil {
		return nil, errors.New("work_items has no master analisa lineage")
	}

	ma, err := s.masterAnalisa.Get(ctx, *analysisMasterID)
	if err != nil {
		return nil, err
	}

	components, err := s.masterAnalisa.ListComponents(ctx, *analysisMasterID)
	if err != nil {
		return nil, err
	}
	var ids []int32
	for _, r := range components {
		if r.ComponentID != nil {
			ids = append(ids, *r.ComponentID)
		}
	}
	live := map[int32]decimal.Decimal{}
	if len(ids) > 0 {
		mhRepo := repository.NewPriceMasterRepo(s.pool)
		live, err = mhRepo.GetMany(ctx, ids)
		if err != nil {
			return nil, err
		}
	}

	var newHS decimal.Decimal
	for _, r := range components {
		hs := decimal.Zero
		if r.ComponentID != nil {
			if lv, ok := live[*r.ComponentID]; ok {
				hs = lv
			}
		}
		compTotal := r.Coefficient.Mul(hs)
		newHS = newHS.Add(compTotal)
	}
	newTotal := newHS.Mul(p.Volume)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	oldTotal := p.TotalCost

	if err := s.work_items.DeleteDetails(ctx, tx, workItemID); err != nil {
		return nil, err
	}

	for _, r := range components {
		hs := decimal.Zero
		if r.ComponentID != nil {
			if lv, ok := live[*r.ComponentID]; ok {
				hs = lv
			}
		}
		compTotal := r.Coefficient.Mul(hs)
		name := ""
		if r.Name != nil {
			name = *r.Name
		}
		sat := ""
		if r.Unit != nil {
			sat = *r.Unit
		}
		maID := *analysisMasterID
		if err := s.work_items.CreateDetail(ctx, tx, repository.CreateDetailInput{
			WorkItemID:       workItemID,
			PriceMasterID:    r.ComponentID,
			AnalysisMasterID: &maID,
			Name:             name,
			Unit:             sat,
			Coefficient:      r.Coefficient,
			UnitPrice:        hs,
			TotalCost:        compTotal.Mul(p.Volume),
			Type:             r.Type,
			SourceCode:       r.ReferenceCode,
		}); err != nil {
			return nil, err
		}
	}

	if err := s.work_items.SetTotal(ctx, tx, workItemID, newHS, newTotal); err != nil {
		return nil, err
	}
	if err := s.work_items.SetBasePrice(ctx, tx, workItemID, ma.UnitPrice); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if s.audit != nil && u != nil {
		pid := p.ProjectID
		pkID := workItemID
		s.audit.Log(audit.Params{
			Action:     audit.ActionRecalculate,
			EntityType: "work_items",
			EntityID:   &workItemID,
			ProjectID:  &pid,
			WorkItemID: &pkID,
			UserID:     u.UserID,
			OldValue:   map[string]any{"totalCost": oldTotal.String()},
			NewValue:   map[string]any{"totalCost": newTotal.String()},
		})
	}
	return s.work_items.Get(ctx, workItemID)
}

// RecalculateAll bulk re-snapshots all AHSP work_items in a project. One txn,
// one audit_log per item ('bulk_recalculate') (§2.4).
func (s *SnapshotService) RecalculateAll(ctx context.Context, projectID int32, u *auth.AuthUser) (int, error) {
	pr := repository.NewWorkItemRepo(s.pool)
	pid := projectID
	items, err := pr.List(ctx, repository.ListWorkItemFilter{ProjectID: &pid})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, p := range items {
		if p.CalculationMethod != models.MethodAHSP {
			continue
		}
		if _, err := s.Recalculate(ctx, p.ID, u); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
