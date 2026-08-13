package handler

import (
	"bytes"
	"context"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

func citem(id int32, cost string, weight float64, hours float64) curveItem {
	return curveItem{id: id, cost: decimal.RequireFromString(cost), weight: weight, hours: hours}
}

// pitem is a planned-series item whose weekly window is set directly (unlike
// citem, whose duration is fed through scheduleCurveItems as hours).
func pitem(id int32, weight, weeks float64) curveItem {
	i := citem(id, "100", weight, 0)
	i.weeks = weeks
	return i
}

func routeCtx(r *http.Request, name, value string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey,
		&chi.Context{URLParams: chi.RouteParams{Keys: []string{name}, Values: []string{value}}}))
}

// --- scheduleCurveItems -----------------------------------------------------

func TestScheduleCurveItemsDurationsFillPlan(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 50, 80),
		citem(2, "100", 50, 40),
	}, 0, 0)
	wantWeeks := []float64{2, 1}
	for i, w := range wantWeeks {
		if items[i].weeks != w {
			t.Errorf("item %d weeks = %v, want %v", i, items[i].weeks, w)
		}
	}
	if items[0].start != 0 || items[1].start != 2 {
		t.Errorf("starts = %v, %v; want 0, 2", items[0].start, items[1].start)
	}
	if cum != 3 {
		t.Errorf("cum = %v, want 3", cum)
	}
}

func TestScheduleCurveItemsTimelineExtendsPlan(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{citem(1, "100", 100, 40)}, 4, 0)
	if items[0].weeks != 1 {
		t.Errorf("weeks = %v, want 1", items[0].weeks)
	}
	if cum != 1 {
		t.Errorf("cum = %v, want 1", cum)
	}
	if items[0].start != 0 {
		t.Errorf("start = %v, want 0", items[0].start)
	}
}

func TestScheduleCurveItemsNoDurationSplitsSlackByCost(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 25, 80),
		citem(2, "300", 75, 0),
	}, 4, 0)
	wantSlack := 4*4.333 - 2
	if diff := math.Abs(items[1].weeks - wantSlack); diff > 0.001 {
		t.Errorf("no-duration weeks = %v, want %v", items[1].weeks, wantSlack)
	}
	if items[0].start != 0 || math.Abs(items[1].start-2) > 0.001 {
		t.Errorf("starts = %v, %v; want 0, 2", items[0].start, items[1].start)
	}
	if math.Abs(cum-4*4.333) > 0.001 {
		t.Errorf("cum = %v, want %v", cum, 4*4.333)
	}
}

func TestScheduleCurveItemsNoDurationCostProportional(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 10, 40),
		citem(2, "100", 30, 0),
		citem(3, "300", 60, 0),
	}, 2, 0)
	wantSlack := 2*4.333 - 1
	if diff := math.Abs(items[1].weeks - wantSlack/4); diff > 0.001 {
		t.Errorf("item2 weeks = %v, want %v", items[1].weeks, wantSlack/4)
	}
	if diff := math.Abs(items[2].weeks - 3*wantSlack/4); diff > 0.001 {
		t.Errorf("item3 weeks = %v, want %v", items[2].weeks, 3*wantSlack/4)
	}
	if math.Abs(cum-2*4.333) > 0.001 {
		t.Errorf("cum = %v, want %v", cum, 2*4.333)
	}
}

func TestScheduleCurveItemsZeroSlackFallsBackToHalfWeek(t *testing.T) {
	items, _ := scheduleCurveItems([]curveItem{
		citem(1, "100", 50, 160),
		citem(2, "100", 50, 0),
	}, 0, 0)
	if items[1].weeks != 0.5 {
		t.Errorf("no-duration weeks = %v, want 0.5", items[1].weeks)
	}
	if items[1].start != 4 {
		t.Errorf("start = %v, want 4", items[1].start)
	}
}

func TestScheduleCurveItemsAllNoDurationProportionalWhenSlack(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 50, 0),
		citem(2, "100", 50, 0),
	}, 0, 0)
	if items[0].weeks != 0.5 || items[1].weeks != 0.5 {
		t.Errorf("weeks = %v, %v; want 0.5, 0.5", items[0].weeks, items[1].weeks)
	}
	if cum != 1 {
		t.Errorf("cum = %v, want 1", cum)
	}
}

func TestScheduleCurveItemsZeroCostItemGetsZeroWeeksNotNaN(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 25, 40),
		citem(2, "0", 0, 0),
		citem(3, "300", 75, 0),
	}, 2, 0)
	if math.IsNaN(items[1].weeks) {
		t.Fatal("zero-cost item produced NaN weeks")
	}
	if items[1].weeks != 0 {
		t.Errorf("zero-cost weeks = %v, want 0", items[1].weeks)
	}
	if items[2].weeks <= 0 {
		t.Errorf("weighted no-duration weeks = %v, want > 0", items[2].weeks)
	}
	if items[1].start != items[2].start {
		t.Errorf("zero-cost start %v != next start %v", items[1].start, items[2].start)
	}
	if cum <= 0 {
		t.Errorf("cum = %v, want > 0", cum)
	}
}

func TestScheduleCurveItemsSubWeekDurations(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "100", 50, 60),
		citem(2, "100", 50, 10),
	}, 0, 0)
	if items[0].weeks != 1.5 || items[1].weeks != 0.25 {
		t.Errorf("weeks = %v, %v; want 1.5, 0.25", items[0].weeks, items[1].weeks)
	}
	if math.Abs(cum-1.75) > 1e-9 {
		t.Errorf("cum = %v, want 1.75", cum)
	}
}

func TestScheduleCurveItemsEveryItemGetsPositiveWeeks(t *testing.T) {
	items, cum := scheduleCurveItems([]curveItem{
		citem(1, "0", 0, 0),
		citem(2, "0", 0, 0),
	}, 0, 0)
	// zero-cost items take the 0.5-week slot: their weeks must still be
	// positive so plannedCurveSeries never divides by zero.
	for i, it := range items {
		if it.weeks <= 0 {
			t.Errorf("item %d weeks = %v, want > 0", i, it.weeks)
		}
	}
	if cum <= 0 {
		t.Errorf("cum = %v, want > 0", cum)
	}
}

// --- curveSpan --------------------------------------------------------------

func TestCurveSpan(t *testing.T) {
	cases := []struct {
		name        string
		cum, rawe   float64
		want        int
	}{
		{"zero", 0, 0, 1},
		{"fractional stays ceil", 0.7, 0, 1},
		{"ceil upper", 2.3, 0, 3},
		{"elapsed dominates", 2, 10.1, 11},
		{"elapsed exact week boundary", 2, 10, 10},
		{"elapsed below cum", 1, 0.5, 1},
		{"negative elapsed ignored", 2, -5, 2},
		{"cap cumulative", 400, 0, sCurveWeekCap},
		{"cap elapsed", 1, 900, sCurveWeekCap},
		{"cap boundary kept", sCurveWeekCap, 0, sCurveWeekCap},
		{"over cap clamps", float64(sCurveWeekCap) + 10, 0, sCurveWeekCap},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := curveSpan(c.cum, c.rawe); got != c.want {
				t.Errorf("curveSpan(%v, %v) = %d, want %d", c.cum, c.rawe, got, c.want)
			}
		})
	}
}

// --- plannedCurveSeries -----------------------------------------------------

func TestPlannedSeriesSingleItemLinear(t *testing.T) {
	items := setStarts([]curveItem{pitem(1, 100, 2)}, 0)
	got := plannedCurveSeries(items, 3)
	want := []int{0, 50, 100}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("week %d = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestPlannedSeriesSequentialItems(t *testing.T) {
	items := setStarts(
		[]curveItem{pitem(1, 40, 1), pitem(2, 60, 1)},
		0, 1)
	got := plannedCurveSeries(items, 3)
	want := []int{0, 40, 100}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("week %d = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestPlannedSeriesNeverExceedsHundred(t *testing.T) {
	// Regression: weights are already percent; scaling them again produced
	// planned values in the thousands.
	items := setStarts(
		[]curveItem{pitem(1, 50, 8), pitem(2, 50, 8)},
		0, 8)
	for span := 2; span <= 30; span++ {
		for _, v := range plannedCurveSeries(items, span) {
			if v < 0 || v > 100 {
				t.Fatalf("planned value %d outside [0,100] (span %d)", v, span)
			}
		}
	}
}

func TestPlannedSeriesMonotonic(t *testing.T) {
	items := setStarts(
		[]curveItem{pitem(1, 25, 1), pitem(2, 25, 2), pitem(3, 50, 0.5)},
		0, 1, 3)
	got := plannedCurveSeries(items, 8)
	prev := 0
	for k, v := range got {
		if v < prev {
			t.Fatalf("planned not monotonic at week %d: %d < %d", k, v, prev)
		}
		prev = v
	}
	if got[len(got)-1] != 100 {
		t.Errorf("endpoint = %d, want 100", got[len(got)-1])
	}
}

func TestPlannedSeriesSpanOneIsZero(t *testing.T) {
	got := plannedCurveSeries([]curveItem{pitem(1, 100, 1)}, 1)
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("got %v, want [0]", got)
	}
}

func TestPlannedSeriesZeroWeightItemContributesNothing(t *testing.T) {
	items := setStarts(
		[]curveItem{pitem(1, 0, 1), pitem(2, 100, 1)},
		0, 0)
	got := plannedCurveSeries(items, 3)
	if got[1] != 100 {
		t.Errorf("week 1 = %d, want 100", got[1])
	}
}

func TestPlannedSeriesItemStartingAfterSpanBounded(t *testing.T) {
	items := setStarts(
		[]curveItem{pitem(1, 60, 0.5), pitem(2, 40, 0.5)},
		0, 500)
	got := plannedCurveSeries(items, 4)
	want := []int{0, 60, 60, 60}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("week %d = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestPlannedSeriesRoundedWeightsSumToHundred(t *testing.T) {
	items := setStarts(
		[]curveItem{pitem(1, 33.33, 1), pitem(2, 33.33, 1), pitem(3, 33.34, 1)},
		0, 1, 2)
	got := plannedCurveSeries(items, 4)
	for i, v := range got {
		if v > 100 || v < 0 {
			t.Errorf("week %d = %d outside [0,100]", i, v)
		}
	}
	if got[1] != 33 || got[3] != 100 {
		t.Errorf("got %v, want week1=33 and endpoint=100", got)
	}
}

func TestPlannedSeriesRandomizedInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for iter := 0; iter < 200; iter++ {
		n := 1 + rng.Intn(12)
		items := make([]curveItem, n)
		total := 0.0
		for i := range items {
			w := rng.Float64() * 40
			items[i] = citem(int32(i+1), "1", w, 40)
			total += w
		}
		if total == 0 {
			total = 1
		}
		scale := 100 / total
		for i := range items {
			items[i].weight = items[i].weight * scale
		}
		// Random starts: derive from fractions of a 30-week plan.
		span := 2 + rng.Intn(30)
		start := 0.0
		for i := range items {
			items[i].start = start
			items[i].weeks = 0.5 + rng.Float64()*6
			start += items[i].weeks * (0.5 + rng.Float64())
		}
		got := plannedCurveSeries(items, span)
		prev := 0
		for k, v := range got {
			if v < 0 || v > 100 {
				t.Fatalf("iter %d: week %d = %d outside [0,100]", iter, k, v)
			}
			if v < prev {
				t.Fatalf("iter %d: not monotonic at week %d (%d < %d)", iter, k, v, prev)
			}
			prev = v
		}
	}
}

func setStarts(items []curveItem, starts ...float64) []curveItem {
	for i, s := range starts {
		items[i].start = s
	}
	return items
}

// --- clamp01 ----------------------------------------------------------------

func TestClamp01(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{-1.5, 0}, {-0.0001, 0}, {0, 0}, {0.5, 0.5}, {1, 1}, {1.0001, 1}, {2, 1},
	}
	for _, c := range cases {
		if got := clamp01(c.in); got != c.want {
			t.Errorf("clamp01(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// --- daysBetween ------------------------------------------------------------

func TestDaysBetween(t *testing.T) {
	loc := time.UTC
	base := time.Date(2024, 2, 10, 15, 0, 0, 0, loc)
	cases := []struct {
		name string
		a, b time.Time
		want int
	}{
		{"same day", base, base, 0},
		{"seven days", base, base.AddDate(0, 0, 7), 7},
		{"month boundary", time.Date(2024, 2, 28, 0, 0, 0, 0, loc), time.Date(2024, 3, 1, 0, 0, 0, 0, loc), 2},
		{"leap year span", time.Date(2024, 2, 28, 0, 0, 0, 0, loc), time.Date(2024, 3, 2, 0, 0, 0, 0, loc), 3},
		{"year boundary", time.Date(2023, 12, 31, 12, 0, 0, 0, loc), time.Date(2024, 1, 1, 12, 0, 0, 0, loc), 1},
		{"backwards negative", time.Date(2024, 1, 10, 0, 0, 0, 0, loc), time.Date(2024, 1, 1, 0, 0, 0, 0, loc), -9},
		{"sub-day truncates", base, base.Add(12 * time.Hour), 0},
		{"one day plus hours", base, base.Add(30 * time.Hour), 1},
		{"hour seconds no drift", time.Date(2024, 5, 1, 23, 30, 0, 0, loc), time.Date(2024, 5, 3, 1, 30, 0, 0, loc), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := daysBetween(c.a, c.b); got != c.want {
				t.Errorf("daysBetween = %d, want %d", got, c.want)
			}
		})
	}
}

// --- invoice math -----------------------------------------------------------

func mkItem(desc, unit string, qty, price string) models.InvoiceItem {
	return models.InvoiceItem{
		Description: desc,
		Unit:        unit,
		Qty:         decimal.RequireFromString(qty),
		UnitPrice:   decimal.RequireFromString(price),
	}
}

func TestInvoiceGrandTotal(t *testing.T) {
	cases := []struct {
		name     string
		items    []models.InvoiceItem
		discount string
		tax      string
		want     string
	}{
		{"empty", nil, "0", "0", "0"},
		{"plain sum", []models.InvoiceItem{mkItem("A", "bh", "2", "100"), mkItem("B", "bh", "1", "50")}, "0", "0", "250"},
		{"discount", []models.InvoiceItem{mkItem("A", "bh", "2", "100")}, "50", "0", "150"},
		{"discount over subtotal clamps to zero", []models.InvoiceItem{mkItem("A", "bh", "2", "100")}, "999", "0", "0"},
		{"tax on base", []models.InvoiceItem{mkItem("A", "bh", "2", "100")}, "0", "11", "222"},
		{"tax after discount", []models.InvoiceItem{mkItem("A", "bh", "2", "100")}, "50", "11", "166.50"},
		{"negative discount inflates", []models.InvoiceItem{mkItem("A", "bh", "2", "100")}, "-25", "0", "225"},
		{"sub-cent total rounds", []models.InvoiceItem{mkItem("A", "bh", "1", "0.01")}, "0", "11", "0.01"},
		{"fractional qty", []models.InvoiceItem{mkItem("A", "m3", "0.5", "3")}, "0", "0", "1.50"},
		{"tax rounds half up", []models.InvoiceItem{mkItem("A", "bh", "1", "0.15")}, "0", "11", "0.17"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := invoiceGrandTotal(c.items, decimal.RequireFromString(c.discount), decimal.RequireFromString(c.tax))
			if !got.Equal(decimal.RequireFromString(c.want)) {
				t.Errorf("grand total = %s, want %s", got.String(), c.want)
			}
		})
	}
}

func TestValidateInvoiceItems(t *testing.T) {
	if err := validateInvoiceItems([]models.InvoiceItem{mkItem("A", "bh", "1", "10")}); err != nil {
		t.Errorf("valid item rejected: %v", err)
	}
	if err := validateInvoiceItems(nil); err != nil {
		t.Errorf("empty list rejected: %v", err)
	}
	// zero unit price is allowed by validation (grand total still floors at 0).
	if err := validateInvoiceItems([]models.InvoiceItem{mkItem("A", "bh", "1", "0")}); err != nil {
		t.Errorf("zero price rejected: %v", err)
	}
	cases := []struct {
		name string
		item models.InvoiceItem
	}{
		{"empty description", mkItem("", "bh", "1", "10")},
		{"zero qty", mkItem("A", "bh", "0", "10")},
		{"negative qty", mkItem("A", "bh", "-1", "10")},
		{"empty unit", mkItem("A", "", "1", "10")},
		{"negative price", mkItem("A", "bh", "1", "-1")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := validateInvoiceItems([]models.InvoiceItem{c.item}); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

// --- small helpers ----------------------------------------------------------

func TestFirstNonNil(t *testing.T) {
	a, b := "a", "b"
	var pn *string
	if got := firstNonNil(pn, nil); got != nil {
		t.Errorf("both nil = %v, want nil", got)
	}
	if got := firstNonNil(nil, &b); got != &b {
		t.Errorf("nil then b = %v", got)
	}
	if got := firstNonNil(&a, &b); got != &a {
		t.Errorf("a then b = %v", got)
	}
	if got := firstNonNil(&a, nil); got != &a {
		t.Errorf("a then nil = %v", got)
	}
}

func TestPtrFromTime(t *testing.T) {
	if got := ptrFromTime(nil); got != nil {
		t.Errorf("nil = %v, want nil", got)
	}
	ts := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	got := ptrFromTime(&ts)
	if got == nil || *got != "2024-01-02" {
		t.Errorf("got %v, want 2024-01-02", got)
	}
}

func TestClampProgress(t *testing.T) {
	cases := []struct{ in, want int }{
		{-1, 0}, {0, 0}, {1, 1}, {50, 50}, {99, 99}, {100, 100}, {101, 100}, {1000, 100},
	}
	for _, c := range cases {
		if got := clampProgress(c.in); got != c.want {
			t.Errorf("clampProgress(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestAtoi32(t *testing.T) {
	cases := []struct {
		in      string
		want    int32
		wantErr bool
	}{
		{"0", 0, false}, {"123", 123, false}, {"-5", -5, false},
		{"2147483647", math.MaxInt32, false}, {"-2147483648", math.MinInt32, false},
		{"2147483648", 0, true}, {"-2147483649", 0, true},
		{"", 0, true}, {"abc", 0, true}, {"12.5", 0, true}, {"12 ", 0, true}, {"+5", 5, false},
	}
	for _, c := range cases {
		got, err := atoi32(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("atoi32(%q) = %v, want error", c.in, got)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("atoi32(%q) = %v, %v; want %v", c.in, got, err, c.want)
		}
	}
}

func TestParseIntParam(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    int32
		wantErr bool
	}{
		{"valid", "42", 42, false},
		{"max int32", "2147483647", math.MaxInt32, false},
		{"zero rejected", "0", 0, true},
		{"negative rejected", "-3", 0, true},
		{"non numeric", "abc", 0, true},
		{"empty", "", 0, true},
		{"overflow", "2147483648", 0, true},
		{"fractional", "1.5", 0, true},
		{"spaces", " 5", 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/x", nil)
			if c.value != "" {
				r = routeCtx(r, "id", c.value)
			}
			w := httptest.NewRecorder()
			got, ok := parseIntParam(w, r, "id")
			if c.wantErr {
				if ok {
					t.Errorf("expected error, got %d", got)
				}
				if w.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want 400", w.Code)
				}
				return
			}
			if !ok || got != c.want {
				t.Errorf("got %d ok=%v, want %d", got, ok, c.want)
			}
			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", w.Code)
			}
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"a":1}`))
	var dst struct{ A int `json:"a"` }
	w := httptest.NewRecorder()
	if !decodeJSON(w, r, &dst) || dst.A != 1 {
		t.Errorf("valid body failed: %+v", dst)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	invalid := []struct {
		name, body string
	}{
		{"empty", ""},
		{"garbage", `{bad json`},
		{"unexpected type", `"string"`},
	}
	for _, c := range invalid {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(c.body))
			w := httptest.NewRecorder()
			if decodeJSON(w, r, &dst) {
				t.Error("expected failure")
			}
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(`"error"`)) {
				t.Errorf("body = %s, want error payload", w.Body.String())
			}
		})
	}
}