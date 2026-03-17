package metrics

import (
	"sync"
	"testing"
)

// TestCacheCH_ConcurrentInit verifies that concurrent calls to CacheCH() when
// Store is nil and ClickHouse env vars are absent all return an error and do
// not race on the Store variable. Run with -race to verify mutex safety.
func TestCacheCH_ConcurrentInit(t *testing.T) {
	// Save and clear the store; restore after the test.
	oldStore := Store
	Store = nil
	defer func() { Store = oldStore }()

	const goroutines = 20
	errs := make([]error, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			errs[i] = CacheCH()
		}()
	}
	wg.Wait()

	// All goroutines must have received an error (missing CLICKHOUSE_HOST etc.)
	for i, err := range errs {
		if err == nil {
			t.Errorf("goroutine %d: expected error from CacheCH() with no env vars, got nil", i)
		}
	}

	// Store must still be nil — no partial initialisation should have occurred.
	if Store != nil {
		t.Errorf("expected Store to remain nil after all CacheCH() calls failed, got %v", Store)
	}
}
