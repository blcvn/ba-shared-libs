package resilience

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half_open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name         string
	maxFailures  int
	resetTimeout time.Duration
	halfOpenMax  int

	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	halfOpenCount int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		halfOpenMax:  3,
		state:        StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	cb.afterRequest(err)

	return err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			log.Printf("[CircuitBreaker:%s] Transitioning to HALF_OPEN", cb.name)
			return nil
		}
		return fmt.Errorf("circuit breaker %s is OPEN", cb.name)

	case StateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenMax {
			return fmt.Errorf("circuit breaker %s is HALF_OPEN (max requests reached)", cb.name)
		}
		cb.halfOpenCount++
		return nil

	case StateClosed:
		return nil
	}

	return nil
}

func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == StateHalfOpen {
			// Failed in half-open state, go back to open
			cb.state = StateOpen
			log.Printf("[CircuitBreaker:%s] Failed in HALF_OPEN, returning to OPEN", cb.name)
		} else if cb.failures >= cb.maxFailures {
			// Too many failures, open the circuit
			cb.state = StateOpen
			log.Printf("[CircuitBreaker:%s] Opening circuit after %d failures", cb.name, cb.failures)
		}
	} else {
		// Success
		if cb.state == StateHalfOpen {
			// Success in half-open, close the circuit
			cb.state = StateClosed
			cb.failures = 0
			log.Printf("[CircuitBreaker:%s] Closing circuit after successful test", cb.name)
		} else if cb.state == StateClosed {
			// Reset failure count on success
			cb.failures = 0
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	RetryableErrors []error
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Retry executes a function with exponential backoff retry
func Retry(ctx context.Context, config *RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, config.RetryableErrors) {
			return err
		}

		// Don't sleep on last attempt
		if attempt < config.MaxAttempts-1 {
			log.Printf("[Retry] Attempt %d/%d failed: %v. Retrying in %v",
				attempt+1, config.MaxAttempts, err, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}

			// Exponential backoff
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("max retry attempts reached: %w", lastErr)
}

func isRetryable(err error, retryableErrors []error) bool {
	if len(retryableErrors) == 0 {
		// If no specific errors defined, retry all errors
		return true
	}

	for _, retryable := range retryableErrors {
		if errors.Is(err, retryable) {
			return true
		}
	}

	return false
}

// Fallback executes a function with a fallback
func Fallback(ctx context.Context, primary func() error, fallback func() error) error {
	err := primary()
	if err != nil {
		log.Printf("[Fallback] Primary failed: %v. Trying fallback", err)
		return fallback()
	}
	return nil
}
