// Package circuitbreaker provides a simple circuit breaker implementation for external service calls.
// It prevents cascading failures by temporarily blocking calls to failing services.
//
// States:
//   - Closed: Normal operation, requests pass through
//   - Open: Service is failing, requests are rejected immediately
//   - HalfOpen: Testing if service has recovered
package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	// StateClosed allows requests to pass through (normal operation).
	StateClosed State = iota
	// StateOpen rejects requests immediately (service is failing).
	StateOpen
	// StateHalfOpen allows a single test request to check if service recovered.
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Config defines the circuit breaker behavior.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int
	// SuccessThreshold is the number of consecutive successes in half-open state before closing.
	SuccessThreshold int
	// Timeout is how long the circuit stays open before moving to half-open.
	Timeout time.Duration
	// Name is used for logging/metrics.
	Name string
	// IsFailure determines if an error should count as a circuit failure.
	// If nil, all errors are considered failures.
	// Return true if the error should trip the circuit breaker.
	// Use this to exclude expected errors like "not found" from tripping the circuit.
	IsFailure func(err error) bool
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig(name string) Config {
	return Config{
		Name:             name,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		IsFailure:        nil, // All errors are failures by default
	}
}

// CircuitBreaker wraps calls to external services with failure protection.
type CircuitBreaker struct {
	config Config

	mu              sync.RWMutex
	state           State
	failures        int
	successes       int
	lastFailure     time.Time
	lastStateChange time.Time
	onStateChange   func(name string, from, to State)
}

// New creates a new circuit breaker with the given configuration.
func New(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// OnStateChange sets a callback for state transitions (useful for logging/metrics).
func (cb *CircuitBreaker) OnStateChange(fn func(name string, from, to State)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute runs the given function if the circuit allows it.
// Returns ErrCircuitOpen if the circuit is open.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("%s: %w", cb.config.Name, ErrCircuitOpen)
	}

	err := fn(ctx)

	// Determine if this error should count as a failure
	isFailure := err != nil
	if err != nil && cb.config.IsFailure != nil {
		isFailure = cb.config.IsFailure(err)
	}

	if isFailure {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// allowRequest checks if a request should be allowed through.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		// Allow one request through for testing
		return true

	default:
		return false
	}
}

// recordFailure records a failed request.
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.successes = 0
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Single failure in half-open returns to open
		cb.transitionTo(StateOpen)
	}
}

// recordSuccess records a successful request.
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++
	cb.failures = 0

	switch cb.state {
	case StateHalfOpen:
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
		}
	}
}

// transitionTo changes the circuit breaker state.
// Must be called while holding the lock.
func (cb *CircuitBreaker) transitionTo(newState State) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	cb.failures = 0
	cb.successes = 0

	// Notify callback
	if cb.onStateChange != nil {
		// Call async to avoid blocking
		go cb.onStateChange(cb.config.Name, oldState, newState)
	}
}

// Stats returns current circuit breaker statistics.
type Stats struct {
	Name            string        `json:"name"`
	State           string        `json:"state"`
	Failures        int           `json:"failures"`
	Successes       int           `json:"successes"`
	LastFailure     time.Time     `json:"last_failure,omitempty"`
	LastStateChange time.Time     `json:"last_state_change"`
	TimeInState     time.Duration `json:"time_in_state"`
}

// Stats returns current statistics for monitoring.
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		Name:            cb.config.Name,
		State:           cb.state.String(),
		Failures:        cb.failures,
		Successes:       cb.successes,
		LastFailure:     cb.lastFailure,
		LastStateChange: cb.lastStateChange,
		TimeInState:     time.Since(cb.lastStateChange),
	}
}
