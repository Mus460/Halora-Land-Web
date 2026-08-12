package audit

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func ipPtr(s string) *string { return &s }

func TestFromRequestStampsIPAndUA(t *testing.T) {
	p := &Params{Action: ActionLogin, UserID: 1}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.9:12345"
	req.Header.Set("User-Agent", "halora-test/1.0")
	p.FromRequest(req)
	if p.IPAddress == nil || *p.IPAddress != "203.0.113.9" {
		t.Errorf("IPAddress = %v", p.IPAddress)
	}
	if p.UserAgent == nil || *p.UserAgent != "halora-test/1.0" {
		t.Errorf("UserAgent = %v", p.UserAgent)
	}
}

func TestFromRequestDoesNotOverwrite(t *testing.T) {
	preset := "1.1.1.1"
	p := &Params{Action: ActionLogin, IPAddress: &preset}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:1"
	p.FromRequest(req)
	if *p.IPAddress != "1.1.1.1" {
		t.Errorf("IPAddress overwritten: %v", p.IPAddress)
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name string
		xff  string
		addr string
		want string
	}{
		{"xff single", "203.0.113.5", "10.0.0.1:80", "203.0.113.5"},
		{"xff multi", "203.0.113.5, 10.0.0.2", "10.0.0.1:80", "203.0.113.5"},
		{"no xff", "", "10.0.0.7:443", "10.0.0.7"},
		{"no xff no port", "", "10.0.0.7", "10.0.0.7"},
	}
	for _, c := range cases {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = c.addr
		if c.xff != "" {
			req.Header.Set("X-Forwarded-For", c.xff)
		}
		if got := clientIP(req); got != c.want {
			t.Errorf("%s: clientIP = %q want %q", c.name, got, c.want)
		}
	}
}

func TestLoggerWritesAndDrains(t *testing.T) {
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	args := make([]any, 11)
	for i := range args {
		args[i] = pgxmock.AnyArg()
	}
	m.ExpectExec(`INSERT INTO audit_log`).WithArgs(args...).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	l := New(m, 4)
	uid := int32(7)
	pid := int32(1)
	l.Log(Params{Action: ActionLogin, EntityType: "auth", UserID: uid, ProjectID: &pid})
	l.Close()
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoggerDropsWhenQueueFull(t *testing.T) {
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	l := New(m, 2)
	for i := 0; i < 100; i++ {
		l.Log(Params{Action: ActionLogin, UserID: 1})
	}
	l.Close()
	// Log must never block: reaching here is the assertion.
}

func TestLoggerLogNeverPanicsAfterClose(t *testing.T) {
	m, _ := pgxmock.NewPool()
	l := New(m, 2)
	l.Close()
	l.Log(Params{Action: ActionLogin, UserID: 1})
}

func TestWriteSwallowsDBErrors(t *testing.T) {
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	args := make([]any, 11)
	for i := range args {
		args[i] = pgxmock.AnyArg()
	}
	m.ExpectExec(`INSERT INTO audit_log`).WithArgs(args...).WillReturnError(context.Canceled)
	l := &Logger{pool: m, ch: make(chan Params, 1)}
	l.write(context.Background(), Params{Action: ActionCreate, UserID: 1})
	// swallowing is the contract — reaching here means no panic
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestParamsFromRequestSideEffectsNil(t *testing.T) {
	p := &Params{Action: ActionDelete, UserID: 2}
	p.FromRequest(httptest.NewRequest("GET", "/", nil))
	if p.IPAddress == nil || p.UserAgent == nil {
		t.Error("expected both stamped")
	}
}
