package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Limiter is an in-memory fixed-window rate limiter (ARCHITECTURE.md §3.8).
// Suitable for single-instance deployments (e.g. a small number of clients).
// Safe for concurrent use.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
}

type entry struct {
	count   int
	resetAt time.Time
}

func New() *Limiter {
	l := &Limiter{entries: make(map[string]*entry)}
	go l.sweep()
	return l
}

// Result of a limit check.
type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	Reset     int64 // unix seconds
}

// Check increments the counter for key and returns whether it's under limit.
// window is the fixed-window duration.
func (l *Limiter) Check(ctx context.Context, key string, limit int, window time.Duration) Result {
	_ = ctx
	now := time.Now()
	reset := now.Add(window).Unix()

	l.mu.Lock()
	e, ok := l.entries[key]
	if !ok || now.After(e.resetAt) {
		e = &entry{count: 0, resetAt: now.Add(window)}
		l.entries[key] = e
	}
	e.count++
	count := e.count
	l.mu.Unlock()

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return Result{Allowed: count <= limit, Limit: limit, Remaining: remaining, Reset: reset}
}

// sweep periodically removes expired entries to bound memory.
func (l *Limiter) sweep() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		l.mu.Lock()
		for k, e := range l.entries {
			if now.After(e.resetAt) {
				delete(l.entries, k)
			}
		}
		l.mu.Unlock()
	}
}

// Middleware rate-limits a request by the given keyFn and limit/window. On 429
// it writes Retry-After and X-RateLimit-* headers (mirrors current contract).
func (l *Limiter) Middleware(keyFn func(*http.Request) string, limit int, window time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := l.Check(r.Context(), keyFn(r), limit, window)
		h := w.Header()
		h.Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
		h.Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(res.Reset, 10))
		if !res.Allowed {
			retry := int(time.Until(time.Unix(res.Reset, 0)).Seconds())
			if retry < 1 {
				retry = 1
			}
			h.Set("Retry-After", strconv.Itoa(retry))
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"Terlalu banyak permintaan. Coba lagi nanti.","retryAfter":%d}`, retry)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ByIP builds a key from "route:ip" using the first X-Forwarded-For hop.
func ByIP(route string) func(*http.Request) string {
	return func(r *http.Request) string {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			for i := 0; i < len(xff); i++ {
				if xff[i] == ',' {
					ip = xff[:i]
					break
				}
			}
			ip = strings.TrimSpace(ip)
		}
		return fmt.Sprintf("%s:%s", route, ip)
	}
}
