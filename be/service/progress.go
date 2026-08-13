package service

import (
	"context"
	"math"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/shopspring/decimal"
)

// WeightedItem is the shared progress/weight view of a work item. Weight is
// the item's cost share of the project subtotal (percent, rounded to 2dp);
// Hours is its estimated duration (volume × duration, 0 when unknown). Both
// feed the S-curve planned line, the monitoring progress page and the
// work-item section tables.
type WeightedItem struct {
	ID          int32
	Category    models.WorkCategory
	Description string
	Volume      string
	Unit        string
	Cost        decimal.Decimal
	Progress    int
	Weight      float64
	Hours       float64
}

// ProgressService is the single source of truth for work-item weights and
// schedule hours. All consumers (S-curve, monitoring, work-item pages) read
// from it so the weight/progress logic is never duplicated.
type ProgressService struct {
	pool database.Pool
}

func NewProgressService(pool database.Pool) *ProgressService {
	return &ProgressService{pool: pool}
}

// Items returns every live work item of the project with its cost, progress,
// cost weight (% of the project subtotal) and duration hours (volume ×
// duration, 0 when unknown), plus the project subtotal. Weights sum to 100
// when the subtotal is positive.
func (s *ProgressService) Items(ctx context.Context, projectID int32) ([]WeightedItem, decimal.Decimal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, category, description, volume::text, unit, "totalCost"::text, progress,
			duration::text
		FROM work_items WHERE "projectId" = $1 AND "deletedAt" IS NULL
		ORDER BY category, id`, projectID)
	if err != nil {
		return nil, decimal.Zero, err
	}
	defer rows.Close()

	items := []WeightedItem{}
	total := decimal.Zero
	for rows.Next() {
		var id int32
		var kat models.WorkCategory
		var description string
		var volStr, unit, costStr, durStr string
		var progress int
		if err := rows.Scan(&id, &kat, &description, &volStr, &unit, &costStr, &progress, &durStr); err != nil {
			return nil, decimal.Zero, err
		}
		cost, err := decimal.NewFromString(costStr)
		if err != nil {
			continue
		}
		it := WeightedItem{ID: id, Category: kat, Description: description, Volume: volStr, Unit: unit, Cost: cost, Progress: progress}
		vol, errV := decimal.NewFromString(volStr)
		dur, errD := decimal.NewFromString(durStr)
		if errV == nil && errD == nil && vol.IsPositive() && dur.IsPositive() {
			it.Hours, _ = vol.Mul(dur).Float64()
		}
		items = append(items, it)
		total = total.Add(cost)
	}
	if err := rows.Err(); err != nil {
		return nil, decimal.Zero, err
	}
	if total.IsPositive() {
		for i := range items {
			f, _ := items[i].Cost.Div(total).Mul(decimal.NewFromInt(100)).Float64()
			items[i].Weight = math.Round(f*100) / 100
		}
	}
	return items, total, nil
}