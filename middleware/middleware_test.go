package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddStandardHttpHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	AddStandardHttpHeaders(rr)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got %q", got)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", got)
	}
	if got := rr.Header().Get("Cache-Control"); got == "" {
		t.Error("expected Cache-Control header to be set")
	}
}

func TestHandlePreflightRequest_OptionsMethod(t *testing.T) {
	req, _ := http.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()
	HandlePreflightRequest(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers header to be set")
	}
}

func TestHandlePreflightRequest_NonOptionsMethod(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	HandlePreflightRequest(rr, req)

	// Non-OPTIONS requests should not set CORS headers
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("expected no Access-Control-Allow-Methods for GET, got %q", got)
	}
}
