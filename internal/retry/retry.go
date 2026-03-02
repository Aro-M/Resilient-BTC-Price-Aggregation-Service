package retry

import (
	"context"
	"time"
)

func Do(ctx context.Context, maxRetries int, fn func() error) (int, error) {
	const base = 500 * time.Millisecond
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return attempt + 1, ctx.Err()
		}
		if err := fn(); err == nil {
			return attempt + 1, nil
		} else {
			lastErr = err
		}
		if attempt == maxRetries {
			break
		}
		exponentialWait := base * time.Duration(1<<uint(attempt))
		select {
		case <-ctx.Done():
			return attempt + 1, ctx.Err()
		case <-time.After(exponentialWait):
		}
	}
	return maxRetries + 1, lastErr
}
