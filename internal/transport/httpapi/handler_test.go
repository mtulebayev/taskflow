package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"taskflow/internal/repository/memory"
	"taskflow/internal/service"
)

type fakeDB struct{ err error }

func (f fakeDB) Ping(context.Context) error { return f.err }

func newTestServer(db Pinger) http.Handler {
	svc := service.NewTaskService(memory.NewTaskRepo())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // в тестах логи не нужны
	return NewHandler(svc, db, logger).Routes()
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestCreateThenGet(t *testing.T) {
	h := newTestServer(fakeDB{})

	rec := do(h, http.MethodPost, "/tasks", `{"title":"buy milk"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /tasks: want 201, got %d", rec.Code)
	}

	rec = do(h, http.MethodGet, "/tasks/1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /tasks/1: want 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "buy milk") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		db     Pinger
		method string
		path   string
		body   string
		want   int
	}{
		{"healthz", fakeDB{}, http.MethodGet, "/healthz", "", http.StatusOK},
		{"readyz db up", fakeDB{}, http.MethodGet, "/readyz", "", http.StatusOK},
		{"readyz db down", fakeDB{err: errors.New("down")}, http.MethodGet, "/readyz", "", http.StatusServiceUnavailable},
		{"invalid id", fakeDB{}, http.MethodGet, "/tasks/abc", "", http.StatusBadRequest},
		{"not found", fakeDB{}, http.MethodGet, "/tasks/999", "", http.StatusNotFound},
		{"empty title", fakeDB{}, http.MethodPost, "/tasks", `{"title":"  "}`, http.StatusBadRequest},
		{"broken json", fakeDB{}, http.MethodPost, "/tasks", `{`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newTestServer(tt.db), tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Fatalf("want %d, got %d (body: %s)", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}
