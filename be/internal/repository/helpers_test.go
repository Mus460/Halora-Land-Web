package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestScanDec(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1000.5", "1000.5"},
		{"0", "0"},
		{"-7.25", "-7.25"},
		{"garbage", "0"}, // unparsable → zero
	}
	for _, c := range cases {
		if got := scanDec(c.in); !got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("scanDec(%q) = %s want %s", c.in, got, c.want)
		}
	}
}

func TestScanDecPtr(t *testing.T) {
	valid := sql.NullString{String: "42.5", Valid: true}
	got := scanDecPtr(valid)
	if got == nil || !got.Equal(decimal.RequireFromString("42.5")) {
		t.Errorf("valid → %v", got)
	}
	if scanDecPtr(sql.NullString{Valid: false}) != nil {
		t.Error("invalid → should be nil")
	}
	if scanDecPtr(sql.NullString{String: "abc", Valid: true}) != nil {
		t.Error("unparsable → should be nil")
	}
}

func TestStrPtr(t *testing.T) {
	v := sql.NullString{String: "x", Valid: true}
	if got := strPtr(v); got == nil || *got != "x" {
		t.Errorf("got %v", got)
	}
	if strPtr(sql.NullString{Valid: false}) != nil {
		t.Error("invalid → nil")
	}
}

func TestTimePtr(t *testing.T) {
	in := sql.NullString{String: "2026-01-02T15:04:05Z", Valid: true}
	got := timePtr(in)
	if got == nil {
		t.Fatal("expected time")
	}
	if !got.Equal(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Errorf("got %v", got)
	}
	if timePtr(sql.NullString{String: "not-a-time", Valid: true}) != nil {
		t.Error("unparsable → nil")
	}
	if timePtr(sql.NullString{Valid: false}) != nil {
		t.Error("invalid → nil")
	}
}

func TestI32Ptr(t *testing.T) {
	got := i32Ptr(sql.NullInt32{Int32: 9, Valid: true})
	if got == nil || *got != 9 {
		t.Errorf("got %v", got)
	}
	if i32Ptr(sql.NullInt32{Valid: false}) != nil {
		t.Error("invalid → nil")
	}
}

func TestDecPtrArg(t *testing.T) {
	if decPtrArg(nil) != nil {
		t.Error("nil → nil")
	}
	d := decimal.RequireFromString("1.5")
	if got := decPtrArg(&d); got != "1.5" {
		t.Errorf("got %v", got)
	}
}

func TestDecArg(t *testing.T) {
	if got := decArg(decimal.RequireFromString("2.75")); got != "2.75" {
		t.Errorf("got %v", got)
	}
}

func TestArgPlaceholder(t *testing.T) {
	for i := 1; i <= 3; i++ {
		if got := argPlaceholder(i); got != string(rune('0'+i)) {
			t.Errorf("argPlaceholder(%d) = %q", i, got)
		}
	}
	if argPlaceholder(12) != "12" {
		t.Error("two-digit placeholder")
	}
}

func TestPad3(t *testing.T) {
	cases := map[int]string{1: "001", 12: "012", 123: "123", 1234: "1234"}
	for in, want := range cases {
		if got := pad3(in); got != want {
			t.Errorf("pad3(%d) = %q want %q", in, got, want)
		}
	}
}
