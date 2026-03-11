package common

import (
	"fmt"
	"net/http"
	"strconv"
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

// ParsePageParams parses ?page and ?page_size query parameters.
// Defaults: page=1, page_size=50. page_size is capped at 500.
func ParsePageParams(r *http.Request) (page, pageSize int, err error) {
	page = 1
	pageSize = 50
	if v := r.URL.Query().Get("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil || page < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		pageSize, err = strconv.Atoi(v)
		if err != nil || pageSize < 1 || pageSize > 500 {
			return 0, 0, fmt.Errorf("page_size must be between 1 and 500")
		}
	}
	return page, pageSize, nil
}
