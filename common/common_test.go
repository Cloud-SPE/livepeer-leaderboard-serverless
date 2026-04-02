package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSince_InvalidFloat(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?since=notanumber", nil)
	_, err := parseSince(req)
	if err == nil {
		t.Error("expected error for non-numeric since parameter")
	}
}

func TestParseUntil_InvalidFloat(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?until=notanumber", nil)
	_, err := parseUntil(req)
	if err == nil {
		t.Error("expected error for non-numeric until parameter")
	}
}

func TestRespondWithError_SetsStatusAndBody(t *testing.T) {
	rr := httptest.NewRecorder()
	RespondWithError(rr, errors.New("something went wrong"), http.StatusBadRequest)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}
	if body["error"] != "something went wrong" {
		t.Errorf("expected error message 'something went wrong', got %q", body["error"])
	}
}

func TestHandleInternalError_Returns500(t *testing.T) {
	rr := httptest.NewRecorder()
	HandleInternalError(rr, errors.New("db failure"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestHandleBadRequest_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()
	HandleBadRequest(rr, errors.New("invalid param"))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}
