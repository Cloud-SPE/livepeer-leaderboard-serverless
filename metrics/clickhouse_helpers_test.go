package metrics

import (
	"testing"
)

func TestEnvOrDefault_ReturnsEnvWhenSet(t *testing.T) {
	t.Setenv("TEST_CH_HELPER_VAR", "custom-value")
	got := envOrDefault("TEST_CH_HELPER_VAR", "fallback")
	if got != "custom-value" {
		t.Errorf("expected 'custom-value', got %q", got)
	}
}

func TestEnvOrDefault_ReturnsFallbackWhenUnset(t *testing.T) {
	// Use a name that will never be set in CI or local environments.
	got := envOrDefault("TEST_CH_DEFINITELY_UNSET_XYZ123", "fallback")
	if got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestEnvOrDefault_ReturnsFallbackWhenBlank(t *testing.T) {
	t.Setenv("TEST_CH_BLANK_VAR", "   ")
	got := envOrDefault("TEST_CH_BLANK_VAR", "fallback")
	if got != "fallback" {
		t.Errorf("expected 'fallback' for whitespace-only value, got %q", got)
	}
}

func TestProtocolFromString_HTTP(t *testing.T) {
	http := protocolFromString("http")
	https := protocolFromString("https")
	unknown := protocolFromString("bogus")

	if http != https {
		t.Error("expected http and https to map to the same protocol")
	}
	if http != unknown {
		t.Error("expected unknown value to default to HTTP protocol")
	}
}

func TestProtocolFromString_Native(t *testing.T) {
	native := protocolFromString("native")
	http := protocolFromString("http")

	if native == http {
		t.Error("expected 'native' to map to a different protocol than 'http'")
	}
}

func TestProtocolFromString_CaseInsensitive(t *testing.T) {
	lower := protocolFromString("native")
	upper := protocolFromString("NATIVE")
	mixed := protocolFromString("Native")

	if lower != upper || lower != mixed {
		t.Error("expected protocolFromString to be case-insensitive")
	}
}
