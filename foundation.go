package foundation

import (
	"context"
	"fmt"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
)

// F is the core interface to Foundation. It builds a linked list of functions to be run
// and provides the ability to run them in sequence, as go routines and blocking functions.
type F interface {
	// Name returns the Name of the F instance given to a Runner.
	Name() string

	// Run runs the given Runners in order. These will block until they have completed running.
	Run(context.Context, ...Runner)

	// Parallel narks the current runner as an asynchronous routine.
	Parallel()

	// On returns an EventHook that allows functions to be exeuted when a specifc event happens.
	On() EventHook

	// Error causes execution to exit immediately unless called from within a clean up function in which case the error
	// will just be logged.
	Error(error)
}

// A Runner runs something.
type Runner interface {
	Run(ctx context.Context, f F)
}

// The RunFunc type is an adapter to allow the use of ordinary functions
// as Runners. If f is a function with the appropriate signature,
// RunFunc(f) is a Runner that calls f.
type RunFunc func(ctx context.Context, f F)

// Run calls fn(ctx, f).
func (fn RunFunc) Run(ctx context.Context, f F) {
	fn(ctx, f)
}

// f is an implementation of foundation.F.
type f struct {
	// If this is a sub function this is the parent.
	parent *f
	// Indicates the function has completed execution.
	signalC chan struct{}
	// Explicitly stop the function and call cleanups which should cause the function to complete.
	stopC chan struct{}
	// Errors that occur during execution of this f are pushed onto this channel
	errC chan error
	// Name of the F
	name string
	// Sub functions that are children of this F.
	subs []*f
	// Guards the fields to prevent race conditions.
	mtx sync.RWMutex
	// Indicates the function has run and has finished execution.
	done atomic.Bool
	// Indicates an explicit stop has been called.
	stopped atomic.Bool
	// Indicates if an error has been encountered.
	erred atomic.Bool
	// Wait group for any go routines we want to wait for before Stop() can exit.
	wg sync.WaitGroup
	// parallelC is a channel closed by Parallal() if the f should be non blocking
	parallelC chan struct{}
	// parallel marks the f as non blocking.
	parallel bool
	// Event hooks to be called when certain events happen.
	hooks *eventHooks
}

// newf constructs a new F.
func newf(name string) *f {
	f := &f{
		signalC:   make(chan struct{}),
		parallelC: make(chan struct{}),
		stopC:     make(chan struct{}),
		errC:      make(chan error),
		subs:      make([]*f, 0),
		name:      name,
		hooks:     newEventHooks(),
	}

	return f
}

// Name returns the Name of F.
func (f *f) Name() string {
	return f.name
}

// Run executes the given run function.
func (f *f) Run(ctx context.Context, runners ...Runner) {
	for _, runner := range runners {
		f.run(ctx, runner)
	}
}

// Go executes the given run function in a go routine. This is useful for running asynchronous processes
// which need to block, for example servers / message consumers.
// Foundation will not exit until all go routines have gracefully exited either naturally or via an explicit
// stop call.
// func (f *f) Go(ctx context.Context, runners ...Runner) {
// 	for _, runner := range runners {
// 		f.run(ctx, runner, true)
// 	}
// }

// Parallel marks this f as a parallel routine. If already marked as parallel this is no-op.
func (f *f) Parallel() {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	// If we are already maked as parallel then do nothing.
	if f.parallel {
		return
	}

	close(f.parallelC)
	f.parallel = true
}

// Error records an error. If being called from a Run function this will stop execution preventing any
// further Run functions from being executed and calling any registered clean up functions before exiting.
// If called from a cleanup function the error will logged and the next cleanup function executed.
func (f *f) Error(err error) {
	if done := f.done.Load(); done {
		return
	}

	// Ensure we do not have a nil error.
	if err == nil {
		return
	}

	// Set error state.
	f.erred.Store(true)

	parent := f.parent

	for {
		// No parent so we are at the root, error state would have already been set.
		if parent == nil {
			break
		}

		// Set the parent error state
		parent.erred.Store(true)
		parent = parent.parent
	}

	// Throw a panic
	//
	// This ensures execution of the current function will stop.
	//
	// This will be caught in the wrapped run function or in the cleanup depending on where the
	// Error() is called from.
	panic(err)
}

// On returns an event hook to add functions which will be called when specific events occur.
func (f *f) On() EventHook {
	return f.hooks
}

func (f *f) stop() {
	// Set stopping state to true, used to prevent further Run functions from being executed.
	f.stopped.Store(true)

	// Call Stop() on sub functions in reverse order so we stop the newest first and the oldest last.
	f.mtx.RLock()
	for i := len(f.subs) - 1; i >= 0; i-- {
		f.subs[i].stop()
	}
	f.mtx.RUnlock()

	// Call stop event hooks
	f.runEventHooks(stopEvent)

	// Wait for signal channel to be closed indicating execution has finished
	// and thereofre we can close error channels.
	<-f.signalC

	// Close error channel causing any go routines listening on it to exit.
	close(f.errC)

	// Wait for routines to exit
	f.wg.Wait()

	// Store done state.
	f.done.Store(true)
}

func (f *f) wait() <-chan struct{} {
	// Create a channel to close once all sub functions are complete.
	ch := make(chan struct{})

	// Create a defer function which closes the returned channel
	// when all sub functions a completed. If there are no sub functions
	// this will close the channel.
	defer func() {
		// We always close the channel when exiting this defer.
		defer close(ch)

		// Obtain read lock
		f.mtx.RLock()

		// Loop over the sub functions.
		for i := range f.subs {
			// Wait for the sub to be done
			<-f.subs[i].wait()
		}

		// Release read lock
		f.mtx.RUnlock()

		// If this is the root function and so do not have a parent we can close our signal channel
		// now as all sub functions are complete.
		if f.parent == nil {
			close(f.signalC)
		}

		// Wait for our function to finish executing
		<-f.signalC
	}()

	return ch
}

// TODO: there is a lot of optimisation to do here and better separation of concerns.
// Will tackle that at a later date.
func (f *f) run(ctx context.Context, runner Runner) {
	// If erred prevent the function from being run.
	if f.erred.Load() || f.done.Load() {
		return
	}

	// Build the name of the new sub f
	f.mtx.RLock()
	name := fmt.Sprintf("%s.%d", f.name, len(f.subs)+1)
	f.mtx.RUnlock()

	// Create a new sub function
	sub := newf(name)
	sub.parent = f

	// Add the below go routine to the wg.
	sub.wg.Add(1)

	// Start a go routine to push errors up to the parent. This will run until the sub error channel is closed
	// explicitly on Stop().
	go func() {
		defer sub.wg.Done()

		for {
			err, ok := <-sub.errC
			if !ok {
				return
			}

			f.errC <- err
		}
	}()

	// Add the new sub function to the list of subs.
	f.mtx.Lock()
	f.subs = append(f.subs, sub)
	f.mtx.Unlock()

	waitC := make(chan struct{})

	// Wrap the function so we can add a defer to know when the functio has completed.
	wrapped := func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				if err, ok := r.(error); ok {
					sub.errC <- RuntimeError{
						Stack: stack,
						Cause: err,
					}
				} else {
					sub.errC <- RuntimeError{
						Stack: stack,
						Cause: PanicError{
							Cause: r,
						},
					}
				}
			}

			// Once the function has completed execution close the signal channel and mark as done.
			sub.mtx.Lock()
			if !sub.done.Load() {
				close(sub.signalC)
			}

			sub.mtx.Unlock()

			close(waitC)

			sub.runEventHooks(doneEvent)
		}()

		runner.Run(ctx, sub)
	}

	// Run the wrapped sub f.
	go wrapped()

	// Wait for the function to either complete or gets marked as a
	// parallel function in which case we do not wait.
	select {
	case <-waitC:
	case <-sub.parallelC:
	}
}

func (f *f) runEventHooks(event eventHook) {
	for hook := range slices.Values(f.hooks.get(event)) {
		f.runEventHook(hook)
	}
}

func (f *f) runEventHook(hook EventHookFunc) {
	defer func() {
		stack := debug.Stack()

		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				f.errC <- CleanupError{
					Stack: stack,
					Cause: err,
				}
			} else {
				f.errC <- CleanupError{
					Stack: stack,
					Cause: PanicError{
						Cause: r,
					},
				}
			}
		}
	}()

	hook()
}
