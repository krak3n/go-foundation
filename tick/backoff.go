package tick

import (
	"context"
	"math/rand/v2"
	"time"
)

// A Backoff returns a wait duration for request retries.
type Backoff interface {
	Wait(ctx context.Context, attempt uint8) time.Duration
}

// The BackoffFunc type is an adapter to allow the use of ordinary functions
// as a Backoff. If f is a function with the appropriate signature,
// BackoffFunc(f) is a Backoff that calls f.
type BackoffFunc func(ctx context.Context, attempt uint8) time.Duration

// Wait calls f(ctx, attempt).
func (f BackoffFunc) Wait(ctx context.Context, attempt uint8) time.Duration {
	return f(ctx, attempt)
}

// A BackoffOption applies common backoff options to backoff configuration.
type BackoffOption interface {
	applyBackoffConfig(cfg *backoffConfig)
}

// The BackoffOptionFunc type is an adapter to allow the use of ordinary functions
// as a BackoffOption. If f is a function with the appropriate signature,
// BackoffOptionFunc(f) is a BackoffOption that calls f.
type BackoffOptionFunc func(cfg *backoffConfig)

func (f BackoffOptionFunc) applyBackoffConfig(cfg *backoffConfig) {
	f(cfg)
}

// BackoffOptions is one or more BackoffOption.
type BackoffOptions []BackoffOption

func (opts BackoffOptions) applyBackoffConfig(cfg *backoffConfig) {
	for _, opt := range opts {
		if opt != nil {
			opt.applyBackoffConfig(cfg)
		}
	}
}

// LinearBackoff is a simple backoff that waits the given wait time in between each attempt.
// To apply jitter use the WithJitter Option, if used will return a LinearBackoff with random jitter
// between each attempt.
func LinearBackoff(wait time.Duration, opts ...BackoffOption) Backoff {
	var cfg backoffConfig

	BackoffOptions(opts).applyBackoffConfig(&cfg)

	if jitter := cfg.jitter; jitter > 0 {
		return BackoffFunc(func(context.Context, uint8) time.Duration {
			return applyJitter(wait, jitter)
		})
	}

	return BackoffFunc(func(context.Context, uint8) time.Duration {
		return wait
	})
}

// BackoffExponential produces a backoff with increasing intervals for each attempt.
// The scalar is multiplied times 2 raised to the current attempt.
// So the first retry with a scalar of 100ms is 100ms, while the 5th attempt would be 1.6s.
// To apply jitter use the WithJitter Option which will apply random jitter to exponential wait duration.
func ExponentialBackoff(scalar time.Duration, opts ...BackoffOption) Backoff {
	var cfg backoffConfig

	BackoffOptions(opts).applyBackoffConfig(&cfg)

	if jitter := cfg.jitter; jitter > 0 {
		return BackoffFunc(func(_ context.Context, attempt uint8) time.Duration {
			return applyJitter(scalar*time.Duration(exponentBase2(attempt)), jitter)
		})
	}

	return BackoffFunc(func(_ context.Context, attempt uint8) time.Duration {
		return scalar * time.Duration(exponentBase2(attempt))
	})
}

// backoffConfig holds backoff configuration that applies to different types of back offs.
type backoffConfig struct {
	jitter float64
}

// applyJitter applies jitter to a given duration.
func applyJitter(d time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (rand.Float64()*2 - 1)

	return time.Duration(float64(d) * (1 + multiplier))
}

// exponentBase2 computes 2^(a-1) where a >= 1. If a is 0, the result is 0.
func exponentBase2(a uint8) uint {
	return (1 << a) >> 1
}
