package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// WithTimeout returns a context that will be canceled after the test timeout.
// If the test timeout given via `go test -timeout <timeout>` is less than the
// required timeout, the test will fail immediately.
// This is to prevent tests that you know require a long time to run from
// starting and then timing out. Might as well fail early.
func WithTimeout(
	t *testing.T,
	ctx context.Context,
	timeout time.Duration,
) context.Context {
	t.Helper()
	deadline, hasDeadline := t.Deadline()
	testTimeout := roundDurationUpToSecond(time.Until(deadline))
	// Is test timeout duration longer than the required timeout.
	// Minus 1 second to avoid rounding errors.
	isTestTimeoutOK := testTimeout >= timeout

	if hasDeadline && !isTestTimeoutOK {
		t.Fatal(
			Callers(),
			fmt.Sprintf(
				"test timeout (%s) less than required timeout (%s)",
				testTimeout,
				timeout,
			),
		)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	t.Cleanup(cancel)

	go func() {
		tickAt := time.Minute * 5
		runningFor := time.Duration(0)

		ticker := time.NewTicker(tickAt)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runningFor += tickAt
				t.Logf(
					"%s: test in progress: running for (%s), timeout in (%s)",
					Callers(),
					runningFor,
					testTimeout-runningFor,
				)
			case <-ctx.Done():
				return
			}
		}
	}()

	return ctx
}

// roundDurationUpToSecond rounds the duration up to the nearest second.
func roundDurationUpToSecond(d time.Duration) time.Duration {
	return time.Duration(d.Seconds())*time.Second + time.Second
}
