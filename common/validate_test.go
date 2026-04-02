package common

import (
	"testing"
)

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
