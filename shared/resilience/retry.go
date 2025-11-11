package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryPolicy defines how retries should be performed
type RetryPolicy struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Multiplier for exponential backoff
	Jitter          bool          // Add random jitter to delays
	RetryableErrors []error       // Specific errors that should trigger retry
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, policy.RetryableErrors) {
			return err // Not retryable
		}

		// Don't sleep after last attempt
		if attempt == policy.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, policy)

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", policy.MaxAttempts, lastErr)
}

// RetryWithResult executes a function with retry logic and returns a result
func RetryWithResult[T any](ctx context.Context, policy RetryPolicy, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Execute the function
		res, err := fn()
		if err == nil {
			return res, nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, policy.RetryableErrors) {
			return result, err // Not retryable
		}

		// Don't sleep after last attempt
		if attempt == policy.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, policy)

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return result, fmt.Errorf("max retry attempts (%d) exceeded: %w", policy.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the next retry with exponential backoff
func calculateDelay(attempt int, policy RetryPolicy) time.Duration {
	// Exponential backoff: initialDelay * (multiplier ^ (attempt - 1))
	delay := float64(policy.InitialDelay) * math.Pow(policy.Multiplier, float64(attempt-1))

	// Cap at max delay
	if delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if policy.Jitter {
		jitter := rand.Float64() * delay * 0.1 // 10% jitter
		delay = delay + jitter - (jitter / 2)   // +/- 5%
	}

	return time.Duration(delay)
}

// isRetryable checks if an error should trigger a retry
func isRetryable(err error, retryableErrors []error) bool {
	if err == nil {
		return false
	}

	// If no specific errors defined, retry all errors
	if len(retryableErrors) == 0 {
		return true
	}

	// Check if error matches any retryable error
	for _, retryableErr := range retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// RetryableHTTPStatusCodes returns common retryable HTTP status codes
func RetryableHTTPStatusCodes() []int {
	return []int{
		408, // Request Timeout
		429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504, // Gateway Timeout
	}
}

// IsRetryableHTTPStatus checks if an HTTP status code is retryable
func IsRetryableHTTPStatus(statusCode int) bool {
	retryable := RetryableHTTPStatusCodes()
	for _, code := range retryable {
		if statusCode == code {
			return true
		}
	}
	return false
}

// ExponentialBackoff returns delays for exponential backoff
func ExponentialBackoff(attempt int, initial, max time.Duration) time.Duration {
	delay := float64(initial) * math.Pow(2.0, float64(attempt-1))
	if delay > float64(max) {
		delay = float64(max)
	}
	return time.Duration(delay)
}

// LinearBackoff returns delays for linear backoff
func LinearBackoff(attempt int, delay time.Duration) time.Duration {
	return time.Duration(attempt) * delay
}

// ConstantBackoff returns a constant delay
func ConstantBackoff(delay time.Duration) time.Duration {
	return delay
}
