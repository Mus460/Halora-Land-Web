package service

import (
	"context"
	"sort"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/repository"
)

// RABService computes the RAB rollup centrally in the BE (ARCHITECTURE.md §3.5,
// §8.3). Rates are configurable (no longer hardcoded 0.10/0.11).
type RABService struct {
	pool         database.Pool
	work_items   *repository.WorkItemRepo
	recaps       *repository.RecapRepo
	overheadRate decimal.Decimal
	ppnRate      decimal.Decimal
}

func NewRABService(pool database.Pool, pr *repository.WorkItemRepo, rr *repository.RecapRepo, overhead, ppn decimal.Decimal) *RABService {
	return &RABService{pool: pool, work_items: pr, recaps: rr, overheadRate: overhead, ppnRate: ppn}
}

type RecapGroup struct {
	Category string          `json:"category"`
	Items    []recapItem     `json:"items"`
	Subtotal decimal.Decimal `json:"subtotal"`
}

type recapItem struct {
	ID            int32            `json:"id"`
	Category      string           `json:"category"`
	Description   string           `json:"description"`
	Volume        decimal.Decimal  `json:"volume"`
	Unit          string           `json:"unit"`
	UnitPrice     decimal.Decimal  `json:"unitPrice"`
	TotalCost     decimal.Decimal  `json:"totalCost"`
	Level         *string          `json:"level"`
	TotalDuration *decimal.Decimal `json:"totalDuration"`
	Duration      *decimal.Decimal `json:"duration"`
}

type RecapResult struct {
	Project          projectSummary             `json:"projects"`
	Grouped          map[string][]recapItem     `json:"grouped"`
	Subtotals        map[string]decimal.Decimal `json:"subtotals"`
	SubtotalDuration map[string]decimal.Decimal `json:"subtotalDuration"`
	Summary          rabSummary                 `json:"summary"`
}

type projectSummary struct {
	ID            int32            `json:"id"`
	Name          string           `json:"name"`
	Location      *string          `json:"location"`
	ContractValue *decimal.Decimal `json:"contractValue"`
}

type rabSummary struct {
	Subtotal           decimal.Decimal `json:"subtotal"`
	Margin             decimal.Decimal `json:"margin"`
	SubtotalWithMargin decimal.Decimal `json:"subtotalWithMargin"`
	Overhead           decimal.Decimal `json:"overhead"`
	Profit             decimal.Decimal `json:"profit"`
	SubtotalBeforeTax  decimal.Decimal `json:"subtotalBeforeTax"`
	PPNPct             decimal.Decimal `json:"ppn"`
	TotalPPN           decimal.Decimal `json:"totalPPN"`
	TotalAkhir         decimal.Decimal `json:"totalFinal"`
	TotalDuration      decimal.Decimal `json:"totalDuration"`
}

// Compute builds the full recaps for a project. Formula (§8.3):
//
//	subtotal = Σ work_items.totalCost
//	subtotalWithMargin = subtotal × (1 + margin/100)
//	overhead = subtotalWithMargin × overheadRate
//	profit = (subtotalWithMargin + overhead) × ... (0 here; margin captures profit)
//	subtotalBeforeTax = subtotalWithMargin + overhead + profit
//	totalPPN = subtotalBeforeTax × ppnRate
//	totalFinal = subtotalBeforeTax + totalPPN
func (s *RABService) Compute(ctx context.Context, projectID int32, projects *repository.SummaryProject) (*RecapResult, error) {
	items, err := s.work_items.List(ctx, repository.ListWorkItemFilter{ProjectID: &projectID})
	if err != nil {
		return nil, err
	}
	margin, err := s.recaps.GetMargin(ctx, projectID)
	if err != nil {
		return nil, err
	}

	grouped := map[string][]recapItem{}
	subtotals := map[string]decimal.Decimal{}
	subtotalDuration := map[string]decimal.Decimal{}
	grandTotal := decimal.Zero
	totalDuration := decimal.Zero
	for _, p := range items {
		it := recapItem{
			ID: p.ID, Category: string(p.Category), Description: p.Description, Volume: p.Volume,
			Unit: p.Unit, UnitPrice: p.UnitPrice, TotalCost: p.TotalCost, Level: p.Level,
			TotalDuration: p.TotalDuration, Duration: p.Duration,
		}
		grouped[string(p.Category)] = append(grouped[string(p.Category)], it)
		subtotals[string(p.Category)] = subtotals[string(p.Category)].Add(p.TotalCost)
		grandTotal = grandTotal.Add(p.TotalCost)
		if p.TotalDuration != nil {
			subtotalDuration[string(p.Category)] = subtotalDuration[string(p.Category)].Add(*p.TotalDuration)
			totalDuration = totalDuration.Add(*p.TotalDuration)
		}
	}

	marginFactor := decimal.NewFromInt(1).Add(margin.Div(decimal.NewFromInt(100)))
	subtotalWithMargin := grandTotal.Mul(marginFactor)
	overhead := subtotalWithMargin.Mul(s.overheadRate)
	profit := decimal.Zero
	subtotalBeforeTax := subtotalWithMargin.Add(overhead).Add(profit)
	totalPPN := subtotalBeforeTax.Mul(s.ppnRate)
	totalFinal := subtotalBeforeTax.Add(totalPPN)

	res := &RecapResult{
		Grouped:          grouped,
		Subtotals:        subtotals,
		SubtotalDuration: subtotalDuration,
		Summary: rabSummary{
			Subtotal:           grandTotal,
			Margin:             margin,
			SubtotalWithMargin: subtotalWithMargin,
			Overhead:           overhead,
			Profit:             profit,
			SubtotalBeforeTax:  subtotalBeforeTax,
			PPNPct:             s.ppnRate.Mul(decimal.NewFromInt(100)),
			TotalPPN:           totalPPN,
			TotalAkhir:         totalFinal,
			TotalDuration:      totalDuration,
		},
	}
	if projects != nil {
		res.Project = projectSummary{
			ID: projects.ID, Name: projects.Name,
			Location: projects.Location, ContractValue: projects.ContractValue,
		}
	}
	return res, nil
}

// SortedCategoryes returns the subtotals map keys in deterministic order.
func SortedCategoryes(m map[string]decimal.Decimal) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
