package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter is a Redis fixed-window rate limiter (ARCHITECTURE.md §3.8).
// Falls back to allow-all if Redis is unavailable, so dev is never blocked.
type Limiter struct {
	rdb redis.Cmdable
}

func New(rdb redis.Cmdable) *Limiter {
	return &Limiter{rdb: rdb}
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
	now := time.Now()
	reset := now.Add(window).Unix()
	if l.rdb == nil {
		return Result{Allowed: true, Limit: limit, Remaining: limit, Reset: reset}
	}
	pipe := l.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, _ = pipe.Exec(ctx)

	count, _ := incr.Result()
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return Result{Allowed: int(count) <= limit, Limit: limit, Remaining: remaining, Reset: reset}
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
