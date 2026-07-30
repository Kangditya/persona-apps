package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type pingFunc func(context.Context) error

func (f pingFunc) PingContext(ctx context.Context) error {
	return f(ctx)
}

func TestHealthAndReadiness(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name        string
		path        string
		ping        pingFunc
		wantCode    int
		wantBody    string
		wantContent string
	}{
		{
			name:        "health does not require the database",
			path:        "/health",
			ping:        func(context.Context) error { return errors.New("not called") },
			wantCode:    http.StatusOK,
			wantBody:    "{\"status\":\"ok\"}\n",
			wantContent: "application/json",
		},
		{
			name:        "ready database reachable",
			path:        "/ready",
			ping:        func(context.Context) error { return nil },
			wantCode:    http.StatusOK,
			wantBody:    "{\"status\":\"ready\"}\n",
			wantContent: "application/json",
		},
		{
			name:        "not ready database unreachable",
			path:        "/ready",
			ping:        func(context.Context) error { return errors.New("database unavailable") },
			wantCode:    http.StatusServiceUnavailable,
			wantBody:    "{\"status\":\"unavailable\"}\n",
			wantContent: "application/json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(":0", test.ping, logger)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			server.Handler.ServeHTTP(response, request)

			if response.Code != test.wantCode {
				t.Fatalf("status code = %d, want %d", response.Code, test.wantCode)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
			if content := response.Header().Get("Content-Type"); content != test.wantContent {
				t.Fatalf("content type = %q, want %q", content, test.wantContent)
			}
		})
	}
}

func TestHealthRejectsUnsupportedMethod(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(":0", pingFunc(func(context.Context) error { return nil }), logger)
	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
