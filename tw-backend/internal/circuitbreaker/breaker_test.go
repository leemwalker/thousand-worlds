package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := New(DefaultConfig("test"))

	// Should start closed
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", cb.State())
	}

	// Successful calls should pass through
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	config := Config{
		Name:             "test",
		FailureThreshold: 3,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
	}
	cb := New(config)

	testErr := errors.New("service unavailable")

	// Cause failures up to threshold
	for i := 0; i < config.FailureThreshold; i++ {
		_ = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	// Circuit should now be open
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen, got %v", cb.State())
	}

	// Next call should be rejected
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	config := Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
	}
	cb := New(config)

	// Open the circuit
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.State())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should be allowed (half-open state)
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error in half-open, got %v", err)
	}

	// Should be back to closed after success
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after recovery, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	config := Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
	}
	cb := New(config)

	// Open the circuit
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Fail again in half-open state
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("still failing")
	})

	// Should be back to open
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	config := Config{
		Name:             "concurrent-test",
		FailureThreshold: 100, // High threshold to avoid opening
		SuccessThreshold: 1,
		Timeout:          time.Second,
	}
	cb := New(config)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}()
	}
	wg.Wait()

	// Should still be closed
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", cb.State())
	}
}

func TestCircuitBreaker_StateChangeCallback(t *testing.T) {
	config := Config{
		Name:             "callback-test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
	}
	cb := New(config)

	var mu sync.Mutex
	var transitions []string

	cb.OnStateChange(func(name string, from, to State) {
		mu.Lock()
		transitions = append(transitions, from.String()+"->"+to.String())
		mu.Unlock()
	})

	// Trigger open
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	// Wait for callback
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(transitions) != 1 || transitions[0] != "closed->open" {
		t.Errorf("expected [closed->open], got %v", transitions)
	}
	mu.Unlock()
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := New(DefaultConfig("stats-test"))

	// Generate some activity
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	stats := cb.Stats()
	if stats.Name != "stats-test" {
		t.Errorf("expected name stats-test, got %s", stats.Name)
	}
	if stats.State != "closed" {
		t.Errorf("expected state closed, got %s", stats.State)
	}
	if stats.Successes != 1 {
		t.Errorf("expected 1 success, got %d", stats.Successes)
	}
}

func TestCircuitBreaker_IsFailurePredicate(t *testing.T) {
	// Define a special "expected" error that should not trip the circuit
	var errExpected = errors.New("expected error, not a failure")

	config := Config{
		Name:             "predicate-test",
		FailureThreshold: 1, // Would normally trip after 1 failure
		SuccessThreshold: 1,
		Timeout:          time.Second,
		IsFailure: func(err error) bool {
			// Only real errors trip the circuit, not expected ones
			return err != nil && !errors.Is(err, errExpected)
		},
	}
	cb := New(config)

	// Return the expected error - should NOT trip the circuit
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return errExpected
	})

	// Error should be returned to caller
	if err != errExpected {
		t.Errorf("expected errExpected, got %v", err)
	}

	// But circuit should still be closed
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed (expected error shouldn't trip), got %v", cb.State())
	}

	// Real errors should still trip the circuit
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("real service error")
	})

	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen (real error should trip), got %v", cb.State())
	}
}
