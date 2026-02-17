package common

import (
	"fmt"
	"strings"
	"time"
)

// ValidateDuration checks that a parsed duration is within [min, max].
// Returns an error suitable for a 400 response if out of bounds.
func ValidateDuration(name string, d time.Duration, min, max time.Duration) error {
	if d < min || d > max {
		return fmt.Errorf("%s must be between %s and %s, got %s", name, min, max, d)
	}
	return nil
}

// ValidateOptionalString trims and length-checks a query parameter.
// Returns the trimmed value and an error if it exceeds maxLen.
func ValidateOptionalString(name, value string, maxLen int) (string, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > maxLen {
		return "", fmt.Errorf("%s exceeds maximum length of %d characters", name, maxLen)
	}
	return trimmed, nil
}
