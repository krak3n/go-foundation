package tick

import (
	"context"
	"sync"
	"time"

	"go.krak3n.io/foundation"
)

// Ticker is a limited subset of F providing ticker functionality.
type Ticker interface {
	// Tick returns the current tick time.
	Tick() time.Time
	// Started returns the time the ticker started ticking.
	Started() time.Time
	// Name returns the name of the ticker from it's underlying F.
	Name() string
	// Stop explicitly stops the ticker and calls any cleanup functions.
	Stop()
	// Error throws a foundation error causing the ticker to stop.
	Error(error)
	// Add an event hook to the ticker
	On() foundation.EventHook
}

// Option configures Runner behaviour.
type Option interface {
	apply(*Runner)
}

// Options is one or more Option.
type Options []Option

func (opts Options) apply(r *Runner) {
	for i := range opts {
		if opt := opts[i]; opt != nil {
			opt.apply(r)
		}
	}
}

// The OptionFunc type is an adapter to allow the use of ordinary functions
// as Options. If f is a function with the appropriate signature,
// OptionFunc(f) is a Option that calls f.
type OptionFunc func(*Runner)

func (f OptionFunc) apply(r *Runner) {
	f(r)
}

// WithUntil sets the maximum number of runs for the ticker. Once this limit is reached the function will
// no longer be executed.
func WithUntil(n uint8) Option {
	return OptionFunc(func(r *Runner) {
		r.maxRunCount = n
	})
}

// A TickFunc is a function called on each tickers tick.
type TickFunc func(ctx context.Context, ticker Ticker)

// Run starts a new linear ticker which will execute the given function on ever tick of the given duration.
// The ticker can be explicitly stopped by calling ticker.Stop() from your tick function.
// The ticked time can be accessed via ticker.Tick() from your tick function.
func Run(ctx context.Context, f foundation.F, d time.Duration, fn TickFunc, opts ...Option) {
	Linear(ctx, f, d, fn, opts...)
}

// Linear starts a new linear ticker which will execute the given function on every tick of the given duration.
// The ticker can be explicitly stopped by calling ticker.Stop() from your tick function.
// The ticked time can be accessed via ticker.Tick() from your tick function.
func Linear(ctx context.Context, f foundation.F, d time.Duration, fn TickFunc, opts ...Option) {
	f.Run(ctx, NewRunner(fn, LinearBackoff(d), opts...))
}

// Expoential starts a new expoential ticker which will execute the given function on every tick.
// The ticker can be explicitly stopped by calling ticker.Stop() from your tick function.
// The ticked time can be accessed via ticker.Tick() from your tick function.
func Exponential(ctx context.Context, f foundation.F, until uint8, scaler time.Duration, fn TickFunc, opts ...Option) {
	var backoff Backoff

	if until == 0 {
		backoff = LinearBackoff(scaler)
	} else {
		backoff = ExponentialBackoff(scaler)
	}

	opts = append(opts, WithUntil(until))

	f.Run(ctx, NewRunner(fn, backoff, opts...))
}

// The Runner type is a foundation.Runner which runs a ticker executing a function on each tick.
type Runner struct {
	tick        time.Time
	started     time.Time
	backoff     Backoff
	backoffOpts []BackoffOption
	f           foundation.F
	fn          TickFunc
	stopC       chan struct{}
	mtx         sync.RWMutex
	stopped     bool
	maxRunCount uint8
	runCount    uint8
	hooks       *eventHooks
}

// NewRunner constructs a new foundation.Runner for running tickers.
// The Runner will execute the given function on every tick of the given duration.
func NewRunner(fn TickFunc, backoff Backoff, opts ...Option) *Runner {
	r := &Runner{
		backoff: backoff,
		fn:      fn,
		stopped: true,
	}

	Options(opts).apply(r)

	return r
}

// Name returns the underlying F's name.
func (r *Runner) Name() string {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.f.Name()
}

// Error calls Error(err) on the underlying F which will cause the ticker to stop and F to exit.
func (r *Runner) Error(err error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	r.f.Error(err)
}

// Tick returns the last tick time.
func (r *Runner) Tick() time.Time {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.tick
}

// Run runs the ticker in parallel.
func (r *Runner) Run(ctx context.Context, f foundation.F) {
	f.Parallel()

	f.On().Stop(func() {
		r.Stop()
	})

	r.mtx.Lock()
	r.f = f
	r.hooks = newEventHooks(f)
	r.mtx.Unlock()

	r.start(ctx)
}

// On returns an EventHookt to add event hook callbacl functions.
func (r *Runner) On() foundation.EventHook {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.hooks
}

// Started returns the start time of the ticker.
func (r *Runner) Started() time.Time {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.started
}

// Stop stop the ticker. No-op if already stopped.
func (r *Runner) Stop() {
	r.mtx.RLock()
	if !r.stopped && r.stopC != nil {
		r.mtx.RUnlock()

		r.mtx.Lock()
		close(r.stopC)
		r.stopped = true
		r.mtx.Unlock()
	} else {
		r.mtx.RUnlock()
	}
}

// Start starts the ticker. No-Op if already started.
func (r *Runner) start(ctx context.Context) {
	// Check if we are stopped.
	r.mtx.RLock()
	if !r.stopped {
		r.mtx.RUnlock()

		return
	} else {
		r.mtx.RUnlock()
	}

	// Save state.
	r.mtx.Lock()
	r.started = time.Now()
	r.stopC = make(chan struct{})
	r.stopped = false
	r.mtx.Unlock()

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-r.stopC
		cancel()
	}()

	defer func() {
		r.Stop()
	}()

	// Tick until told to stop.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.mtx.RLock()
			count := r.runCount + 1

			if r.maxRunCount > 0 {
				if count > r.maxRunCount {
					r.mtx.RUnlock()
					return
				}
			}

			r.mtx.RUnlock()

			if err := wait(ctx, count, r.backoff); err != nil {
				return
			}

			r.mtx.Lock()
			r.tick = time.Now()
			r.runCount = count
			r.mtx.Unlock()

			r.fn(ctx, r)
		}
	}
}

// Wait calculates the backoff wait duration based on the attempt number and Backoff given
func wait(ctx context.Context, count uint8, backoff Backoff) error {
	wait := backoff.Wait(ctx, count)

	if wait > 0 {
		timer := time.NewTimer(wait)

		select {
		case <-ctx.Done():

			if !timer.Stop() {
				<-timer.C
			}

			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}

	return nil
}
