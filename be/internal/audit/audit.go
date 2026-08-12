package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/halora-land/halora-be/internal/database"
)

// Action constants (collapses the two parallel TS modules into one set, §3.6).
const (
	ActionCreate          = "CREATE"
	ActionUpdate          = "UPDATE"
	ActionDelete          = "DELETE"
	ActionLogin           = "LOGIN"
	ActionLogout          = "LOGOUT"
	ActionRegister        = "REGISTER"
	ActionExport          = "EXPORT"
	ActionRecalculate     = "recalculate"
	ActionBulkRecalculate = "bulk_recalculate"
)

// Params describes one audit entry to record.
type Params struct {
	Action      string
	EntityType  string
	EntityID    *int32
	ProjectID   *int32
	WorkItemID  *int32
	UserID      int32
	OldValue    any
	NewValue    any
	Description *string
	IPAddress   *string
	UserAgent   *string
}

// Logger queues audit writes to a background worker (true non-blocking —
// survives request cancellation). ARCHITECTURE.md §3.6 porting note.
type Logger struct {
	pool   database.Pool
	ch     chan Params
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

// New creates a Logger with a buffered queue of the given size and starts the
// background worker. Call Close during shutdown to drain.
func New(pool database.Pool, bufferSize int) *Logger {
	l := &Logger{pool: pool, ch: make(chan Params, bufferSize)}
	l.wg.Add(1)
	go l.run()
	return l
}

// Log enqueues an audit entry. Never blocks (drops on full queue or after
// Close so audit never breaks user flows — mirrors current "swallow errors"
// contract, §3.6).
func (l *Logger) Log(p Params) {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	select {
	case l.ch <- p:
	default:
	}
}

// FromRequest stamps IP/User-Agent from the request when not already set.
func (p *Params) FromRequest(r *http.Request) {
	if p.IPAddress == nil {
		ip := clientIP(r)
		p.IPAddress = &ip
	}
	if p.UserAgent == nil {
		ua := r.UserAgent()
		p.UserAgent = &ua
	}
}

func (l *Logger) run() {
	defer l.wg.Done()
	for p := range l.ch {
		l.write(context.Background(), p)
	}
}

// Close drains the queue and stops the worker.
func (l *Logger) Close() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	l.mu.Unlock()
	close(l.ch)
	l.wg.Wait()
}

func (l *Logger) write(ctx context.Context, p Params) {
	var oldB, newB []byte
	if p.OldValue != nil {
		oldB, _ = json.Marshal(p.OldValue)
	}
	if p.NewValue != nil {
		newB, _ = json.Marshal(p.NewValue)
	}
	_, err := l.pool.Exec(ctx, `
		INSERT INTO audit_log ("projectId", "workItemId", "userId", action, "entityType", "entityId", "oldValue", "newValue", description, "ipAddress", "userAgent")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		p.ProjectID, p.WorkItemID, p.UserID, p.Action, p.EntityType, p.EntityID,
		oldB, newB, p.Description, p.IPAddress, p.UserAgent)
	if err != nil {
		// swallow — audit must not break user flows (§3.6)
		_ = err
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}
