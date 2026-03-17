package common

import (
	"net/http"
	"testing"
	"time"
)

// --- ValidateDuration ---

func TestValidateDuration_WithinRange(t *testing.T) {
	if err := ValidateDuration("x", 5*time.Minute, time.Minute, time.Hour); err != nil {
		t.Errorf("expected no error for in-range duration, got: %v", err)
	}
}

func TestValidateDuration_AtMin(t *testing.T) {
	if err := ValidateDuration("x", time.Minute, time.Minute, time.Hour); err != nil {
		t.Errorf("expected no error at min boundary, got: %v", err)
	}
}

func TestValidateDuration_AtMax(t *testing.T) {
	if err := ValidateDuration("x", time.Hour, time.Minute, time.Hour); err != nil {
		t.Errorf("expected no error at max boundary, got: %v", err)
	}
}

func TestValidateDuration_BelowMin(t *testing.T) {
	if err := ValidateDuration("x", 30*time.Second, time.Minute, time.Hour); err == nil {
		t.Error("expected error for below-min duration")
	}
}

func TestValidateDuration_AboveMax(t *testing.T) {
	if err := ValidateDuration("x", 2*time.Hour, time.Minute, time.Hour); err == nil {
		t.Error("expected error for above-max duration")
	}
}

// --- ValidateOptionalString ---

func TestValidateOptionalString_Empty(t *testing.T) {
	got, err := ValidateOptionalString("f", "", 10)
	if err != nil {
		t.Errorf("unexpected error for empty string: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestValidateOptionalString_WithinLimit(t *testing.T) {
	got, err := ValidateOptionalString("f", "hello", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestValidateOptionalString_ExceedsLimit(t *testing.T) {
	_, err := ValidateOptionalString("f", "toolongvalue", 5)
	if err == nil {
		t.Error("expected error for string exceeding max length")
	}
}

func TestValidateOptionalString_TrimsWhitespace(t *testing.T) {
	got, err := ValidateOptionalString("f", "  hi  ", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != "hi" {
		t.Errorf("expected trimmed 'hi', got %q", got)
	}
}

// --- ParsePageParams ---

func TestParsePageParams_Defaults(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	page, pageSize, err := ParsePageParams(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 1 {
		t.Errorf("expected default page=1, got %d", page)
	}
	if pageSize != 50 {
		t.Errorf("expected default page_size=50, got %d", pageSize)
	}
}

func TestParsePageParams_ValidValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page=3&page_size=100", nil)
	page, pageSize, err := ParsePageParams(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 3 {
		t.Errorf("expected page=3, got %d", page)
	}
	if pageSize != 100 {
		t.Errorf("expected page_size=100, got %d", pageSize)
	}
}

func TestParsePageParams_PageZeroReturnsError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page=0", nil)
	if _, _, err := ParsePageParams(req); err == nil {
		t.Error("expected error for page=0")
	}
}

func TestParsePageParams_PageSizeZeroReturnsError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page_size=0", nil)
	if _, _, err := ParsePageParams(req); err == nil {
		t.Error("expected error for page_size=0")
	}
}

func TestParsePageParams_PageSizeAboveMaxReturnsError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page_size=501", nil)
	if _, _, err := ParsePageParams(req); err == nil {
		t.Error("expected error for page_size=501")
	}
}

func TestParsePageParams_PageSizeAtMaxAccepted(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page_size=500", nil)
	_, pageSize, err := ParsePageParams(req)
	if err != nil {
		t.Fatalf("expected no error for page_size=500, got: %v", err)
	}
	if pageSize != 500 {
		t.Errorf("expected page_size=500, got %d", pageSize)
	}
}

func TestParsePageParams_NonIntegerPageReturnsError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?page=abc", nil)
	if _, _, err := ParsePageParams(req); err == nil {
		t.Error("expected error for non-integer page")
	}
}

// --- ParseDurationParam ---

func TestParseDurationParam_Absent(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	got, err := ParseDurationParam(req, "interval", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 15*time.Minute {
		t.Errorf("expected default 15m, got %v", got)
	}
}

func TestParseDurationParam_GoDurationString(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?interval=2h", nil)
	got, err := ParseDurationParam(req, "interval", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2*time.Hour {
		t.Errorf("expected 2h, got %v", got)
	}
}

func TestParseDurationParam_NumericSeconds(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?interval=3600", nil)
	got, err := ParseDurationParam(req, "interval", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != time.Hour {
		t.Errorf("expected 1h from 3600 seconds, got %v", got)
	}
}

func TestParseDurationParam_InvalidValue(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?interval=notanumber", nil)
	if _, err := ParseDurationParam(req, "interval", 15*time.Minute); err == nil {
		t.Error("expected error for invalid duration value")
	}
}

// --- EnvOrDefault ---

func TestEnvOrDefault_StringDefaultReturnsEnvValue(t *testing.T) {
	t.Setenv("TEST_COMMON_ENV_STR", "from-env")
	got := EnvOrDefault("TEST_COMMON_ENV_STR", "default").(string)
	if got != "from-env" {
		t.Errorf("expected 'from-env', got %q", got)
	}
}

func TestEnvOrDefault_StringDefaultReturnsFallback(t *testing.T) {
	got := EnvOrDefault("TEST_COMMON_UNSET_XYZ999", "default").(string)
	if got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestEnvOrDefault_IntDefaultReturnsEnvValue(t *testing.T) {
	t.Setenv("TEST_COMMON_ENV_INT", "42")
	got := EnvOrDefault("TEST_COMMON_ENV_INT", 0).(int)
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestEnvOrDefault_IntDefaultReturnsFallbackOnInvalidValue(t *testing.T) {
	t.Setenv("TEST_COMMON_ENV_INT_BAD", "not-an-int")
	got := EnvOrDefault("TEST_COMMON_ENV_INT_BAD", 7).(int)
	if got != 7 {
		t.Errorf("expected fallback 7 for non-integer env value, got %d", got)
	}
}
