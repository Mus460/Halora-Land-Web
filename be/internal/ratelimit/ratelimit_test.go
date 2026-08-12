package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestCheckUnderLimit(t *testing.T) {
	l := New()
	r := l.Check(t.Context(), "ip", 3, time.Minute)
	if !r.Allowed {
		t.Fatal("first request should be allowed")
	}
	if r.Remaining != 2 {
		t.Errorf("Remaining = %d want 2", r.Remaining)
	}
	if r.Limit != 3 {
		t.Errorf("Limit = %d want 3", r.Limit)
	}
	if r.Reset <= time.Now().Unix() {
		t.Errorf("Reset = %d should be in the future", r.Reset)
	}
}

func TestCheckAtLimit(t *testing.T) {
	l := New()
	var res Result
	for i := 0; i < 3; i++ {
		res = l.Check(t.Context(), "ip", 3, time.Minute)
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if res.Remaining != 0 {
		t.Errorf("Remaining = %d want 0", res.Remaining)
	}
}

func TestCheckOverLimit(t *testing.T) {
	l := New()
	for i := 0; i < 4; i++ {
		l.Check(t.Context(), "ip", 3, time.Minute)
	}
	res := l.Check(t.Context(), "ip", 3, time.Minute)
	if res.Allowed {
		t.Fatal("4th+ request should be blocked")
	}
	if res.Remaining != 0 {
		t.Errorf("Remaining = %d want 0", res.Remaining)
	}
}

func TestWindowReset(t *testing.T) {
	l := New()
	for i := 0; i < 2; i++ {
		l.Check(t.Context(), "ip", 2, 30*time.Millisecond)
	}
	if res := l.Check(t.Context(), "ip", 2, 30*time.Millisecond); res.Allowed {
		t.Fatal("expected blocked before window reset")
	}
	time.Sleep(50 * time.Millisecond)
	res := l.Check(t.Context(), "ip", 2, 30*time.Millisecond)
	if !res.Allowed {
		t.Fatal("expected allowed after window reset")
	}
	if res.Remaining != 1 {
		t.Errorf("Remaining = %d want 1", res.Remaining)
	}
}

func TestKeysAreIsolated(t *testing.T) {
	l := New()
	l.Check(t.Context(), "a", 1, time.Minute)
	if res := l.Check(t.Context(), "a", 1, time.Minute); res.Allowed {
		t.Fatal("key a should be exhausted")
	}
	if res := l.Check(t.Context(), "b", 1, time.Minute); !res.Allowed {
		t.Fatal("key b should be unaffected")
	}
}

func TestMiddlewareAllowsAndBlocks(t *testing.T) {
	l := New()
	var hits int
	h := l.Middleware(
		func(r *http.Request) string { return "ip:1.2.3.4" },
		2, time.Minute,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++ }),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("code = %d want 200", rec.Code)
	}
	if got := rec.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Errorf("X-RateLimit-Limit = %q", got)
	}
	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "1" {
		t.Errorf("X-RateLimit-Remaining = %q", got)
	}
	if rec.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("X-RateLimit-Reset missing")
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest("GET", "/", nil))
	if rec2.Code != http.StatusOK {
		t.Errorf("code = %d want 200", rec2.Code)
	}

	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, httptest.NewRequest("GET", "/", nil))
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("code = %d want 429", rec3.Code)
	}
	if got := rec3.Header().Get("Retry-After"); got == "" {
		t.Error("Retry-After missing on 429")
	}
	if got := rec3.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("X-RateLimit-Remaining = %q want 0", got)
	}
	if hits != 2 {
		t.Errorf("handler hits = %d want 2", hits)
	}
}

func TestByIPUsesRemoteAddr(t *testing.T) {
	keyFn := ByIP("route")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.9:51234"
	if got := keyFn(r); got != "route:203.0.113.9:51234" {
		t.Errorf("key = %q", got)
	}
}

func TestByIPUsesXForwardedFor(t *testing.T) {
	keyFn := ByIP("route")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "proxy:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	if got := keyFn(r); got != "route:203.0.113.5" {
		t.Errorf("key = %q want first XFF hop", got)
	}
}

func TestByIPEmptyXFF(t *testing.T) {
	keyFn := ByIP("route")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.7:99"
	r.Header.Set("X-Forwarded-For", "")
	if got := keyFn(r); got != "route:10.0.0.7:99" {
		t.Errorf("key = %q", got)
	}
}

func TestResetParsableAsUnixSeconds(t *testing.T) {
	l := New()
	res := l.Check(t.Context(), "ip", 5, time.Hour)
	if _, err := strconv.ParseInt(strconv.FormatInt(res.Reset, 10), 10, 64); err != nil {
		t.Errorf("Reset not a unix timestamp: %v", err)
	}
}
