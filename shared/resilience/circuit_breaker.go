package resilience

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota // Normal operation
	StateOpen                // Circuit is open, requests fail fast
	StateHalfOpen            // Testing if service recovered
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

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are being processed
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	maxRequests    uint32        // Max requests allowed in half-open state
	interval       time.Duration // Time window for error rate calculation
	timeout        time.Duration // Time to wait before transitioning from open to half-open
	threshold      uint32        // Minimum number of requests before checking error rate
	failureRate    float64       // Maximum acceptable failure rate (0.0 to 1.0)
	onStateChange  func(from, to State)
	mu             sync.Mutex
	state          State
	counts         *counts
	expiry         time.Time
}

type counts struct {
	requests       uint32
	totalSuccesses uint32
	totalFailures  uint32
	consecutiveSuccesses uint32
	consecutiveFailures  uint32
}

// Settings configures a circuit breaker
type Settings struct {
	Name          string
	MaxRequests   uint32        // Default: 1
	Interval      time.Duration // Default: 60s
	Timeout       time.Duration // Default: 60s
	Threshold     uint32        // Default: 5
	FailureRate   float64       // Default: 0.5 (50%)
	OnStateChange func(name string, from State, to State)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(settings Settings) *CircuitBreaker {
	if settings.MaxRequests == 0 {
		settings.MaxRequests = 1
	}
	if settings.Interval == 0 {
		settings.Interval = 60 * time.Second
	}
	if settings.Timeout == 0 {
		settings.Timeout = 60 * time.Second
	}
	if settings.Threshold == 0 {
		settings.Threshold = 5
	}
	if settings.FailureRate == 0 {
		settings.FailureRate = 0.5
	}

	cb := &CircuitBreaker{
		name:        settings.Name,
		maxRequests: settings.MaxRequests,
		interval:    settings.Interval,
		timeout:     settings.Timeout,
		threshold:   settings.Threshold,
		failureRate: settings.FailureRate,
	}

	if settings.OnStateChange != nil {
		cb.onStateChange = func(from, to State) {
			settings.OnStateChange(settings.Name, from, to)
		}
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given function if the circuit breaker is closed or half-open
func (cb *CircuitBreaker) Execute(fn func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			cb.afterRequest(generation, false)
			panic(r)
		}
	}()

	err = fn()
	cb.afterRequest(generation, err == nil)
	return err
}

// Call is a convenience wrapper for Execute
func (cb *CircuitBreaker) Call(fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := cb.Execute(func() error {
		var fnErr error
		result, fnErr = fn()
		return fnErr
	})
	return result, err
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitOpen
	}

	if state == StateHalfOpen && cb.counts.requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.requests++
	return generation, nil
}

func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if generation != before {
		return // Request was from previous generation
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.totalSuccesses++
	cb.counts.consecutiveSuccesses++
	cb.counts.consecutiveFailures = 0

	if state == StateHalfOpen {
		// If we get enough consecutive successes in half-open, close the circuit
		if cb.counts.consecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.totalFailures++
	cb.counts.consecutiveFailures++
	cb.counts.consecutiveSuccesses = 0

	switch state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		// Any failure in half-open immediately opens the circuit
		cb.setState(StateOpen, now)
	}
}

func (cb *CircuitBreaker) shouldOpen() bool {
	counts := cb.counts

	// Need minimum number of requests before opening
	if counts.requests < cb.threshold {
		return false
	}

	// Calculate failure rate
	rate := float64(counts.totalFailures) / float64(counts.requests)
	return rate >= cb.failureRate
}

func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}

	return cb.state, uint64(cb.counts.requests)
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(prev, state)
	}
}

func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.counts = &counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns statistics about the circuit breaker
func (cb *CircuitBreaker) Counts() (requests, successes, failures uint32) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.counts.requests, cb.counts.totalSuccesses, cb.counts.totalFailures
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.toNewGeneration(time.Now())
	cb.state = StateClosed
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return fmt.Sprintf(
		"CircuitBreaker{name=%s, state=%s, requests=%d, successes=%d, failures=%d}",
		cb.name,
		cb.state.String(),
		cb.counts.requests,
		cb.counts.totalSuccesses,
		cb.counts.totalFailures,
	)
}
