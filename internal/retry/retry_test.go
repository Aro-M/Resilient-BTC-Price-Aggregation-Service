package retry

import (
	"context"
	"errors"
	"testing"
)

func TestDo_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()
	attempts, err := Do(ctx, 3, func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	var calls int

	attempts, err := Do(ctx, 3, func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if calls != 3 {
		t.Fatalf("expected 3 function calls, got %d", calls)
	}
}

func TestDo_ExhaustsMaxRetries(t *testing.T) {
	ctx := context.Background()
	var calls int
	expectedErr := errors.New("persistent error")

	attempts, err := Do(ctx, 2, func() error {
		calls++
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if calls != 3 {
		t.Fatalf("expected 3 function calls, got %d", calls)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int

	attempts, err := Do(ctx, 5, func() error {
		calls++
		if calls == 2 {
			cancel()
		}
		return errors.New("simulated failure")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
	if attempts > 2 {
		t.Fatalf("should have bailed early, got %d attempts", attempts)
	}
}
