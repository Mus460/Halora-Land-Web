package service

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/repository"
)

// RABService computes the RAB rollup centrally in the BE (ARCHITECTURE.md §3.5,
// §8.3). Rates are configurable (no longer hardcoded 0.10/0.11).
type RABService struct {
	pool          *pgxpool.Pool
	pekerjaan     *repository.PekerjaanRepo
	rekap         *repository.RekapRepo
	overheadRate  decimal.Decimal
	ppnRate       decimal.Decimal
}

func NewRABService(pool *pgxpool.Pool, pr *repository.PekerjaanRepo, rr *repository.RekapRepo, overhead, ppn decimal.Decimal) *RABService {
	return &RABService{pool: pool, pekerjaan: pr, rekap: rr, overheadRate: overhead, ppnRate: ppn}
}

type RekapGroup struct {
	Kategori string             `json:"kategori"`
	Items    []rekapItem        `json:"items"`
	Subtotal decimal.Decimal    `json:"subtotal"`
}

type rekapItem struct {
	ID              int32            `json:"id"`
	Kategori        string           `json:"kategori"`
	UraianPekerjaan string           `json:"uraianPekerjaan"`
	Volume          decimal.Decimal  `json:"volume"`
	Satuan          string           `json:"satuan"`
	HargaSatuan     decimal.Decimal  `json:"hargaSatuan"`
	TotalBiaya      decimal.Decimal  `json:"totalBiaya"`
	LevelPekerjaan  *string          `json:"levelPekerjaan"`
	TotalWaktu      *decimal.Decimal `json:"totalWaktu"`
	Waktu           *decimal.Decimal `json:"waktu"`
}

type RekapResult struct {
	Proyek        proyekSummary              `json:"proyek"`
	Grouped       map[string][]rekapItem     `json:"grouped"`
	Subtotals     map[string]decimal.Decimal `json:"subtotals"`
	SubtotalWaktu map[string]decimal.Decimal `json:"subtotalWaktu"`
	Summary       rabSummary                 `json:"summary"`
}

type proyekSummary struct {
	ID          int32            `json:"id"`
	NamaProyek  string           `json:"namaProyek"`
	Lokasi      *string          `json:"lokasi"`
	NilaiKontrak *decimal.Decimal `json:"nilaiKontrak"`
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
	TotalAkhir         decimal.Decimal `json:"totalAkhir"`
	TotalWaktu         decimal.Decimal `json:"totalWaktu"`
}

// Compute builds the full rekap for a project. Formula (§8.3):
//   subtotal = Σ pekerjaan.totalBiaya
//   subtotalWithMargin = subtotal × (1 + margin/100)
//   overhead = subtotalWithMargin × overheadRate
//   profit = (subtotalWithMargin + overhead) × ... (0 here; margin captures profit)
//   subtotalBeforeTax = subtotalWithMargin + overhead + profit
//   totalPPN = subtotalBeforeTax × ppnRate
//   totalAkhir = subtotalBeforeTax + totalPPN
func (s *RABService) Compute(ctx context.Context, proyekID int32, proyek *repository.SummaryProyek) (*RekapResult, error) {
	items, err := s.pekerjaan.List(ctx, repository.ListPekerjaanFilter{ProyekID: &proyekID})
	if err != nil {
		return nil, err
	}
	margin, err := s.rekap.GetMargin(ctx, proyekID)
	if err != nil {
		return nil, err
	}

	grouped := map[string][]rekapItem{}
	subtotals := map[string]decimal.Decimal{}
	subtotalWaktu := map[string]decimal.Decimal{}
	grandTotal := decimal.Zero
	totalWaktu := decimal.Zero
	for _, p := range items {
		it := rekapItem{
			ID: p.ID, Kategori: string(p.Kategori), UraianPekerjaan: p.UraianPekerjaan, Volume: p.Volume,
			Satuan: p.Satuan, HargaSatuan: p.HargaSatuan, TotalBiaya: p.TotalBiaya, LevelPekerjaan: p.LevelPekerjaan,
			TotalWaktu: p.TotalWaktu, Waktu: p.Waktu,
		}
		grouped[string(p.Kategori)] = append(grouped[string(p.Kategori)], it)
		subtotals[string(p.Kategori)] = subtotals[string(p.Kategori)].Add(p.TotalBiaya)
		grandTotal = grandTotal.Add(p.TotalBiaya)
		if p.TotalWaktu != nil {
			subtotalWaktu[string(p.Kategori)] = subtotalWaktu[string(p.Kategori)].Add(*p.TotalWaktu)
			totalWaktu = totalWaktu.Add(*p.TotalWaktu)
		}
	}

	marginFactor := decimal.NewFromInt(1).Add(margin.Div(decimal.NewFromInt(100)))
	subtotalWithMargin := grandTotal.Mul(marginFactor)
	overhead := subtotalWithMargin.Mul(s.overheadRate)
	profit := decimal.Zero
	subtotalBeforeTax := subtotalWithMargin.Add(overhead).Add(profit)
	totalPPN := subtotalBeforeTax.Mul(s.ppnRate)
	totalAkhir := subtotalBeforeTax.Add(totalPPN)

	res := &RekapResult{
		Grouped:        grouped,
		Subtotals:      subtotals,
		SubtotalWaktu:  subtotalWaktu,
		Summary: rabSummary{
			Subtotal:           grandTotal,
			Margin:             margin,
			SubtotalWithMargin: subtotalWithMargin,
			Overhead:           overhead,
			Profit:             profit,
			SubtotalBeforeTax:  subtotalBeforeTax,
			PPNPct:             s.ppnRate.Mul(decimal.NewFromInt(100)),
			TotalPPN:           totalPPN,
			TotalAkhir:         totalAkhir,
			TotalWaktu:         totalWaktu,
		},
	}
	if proyek != nil {
		res.Proyek = proyekSummary{
			ID: proyek.ID, NamaProyek: proyek.NamaProyek,
			Lokasi: proyek.Lokasi, NilaiKontrak: proyek.NilaiKontrak,
		}
	}
	return res, nil
}

// SortedKategories returns the subtotals map keys in deterministic order.
func SortedKategories(m map[string]decimal.Decimal) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
