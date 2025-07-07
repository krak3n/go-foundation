package foundation

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Run runs a the given foundation runner.
func Run(name string, runner Runner) {
	ctx := context.Background()

	// Initialise new foundation with the given service name.
	f := newf(name)

	// Exit code to use on exit when call os.Exit. 0 indicates success, any other value indicates error.
	var exitCode int

	// Create a wait group to ensure all go routines exit.
	var wg sync.WaitGroup

	// Channels to manage orchestration
	done := make(chan struct{})
	errd := make(chan struct{})

	// Add the two go routines to the wait group.
	wg.Add(2)

	// Start a go routine which reads from the f error channel.
	// If an error is encountered we close the errd channel causing Stop() to be called.
	go func() {
		defer wg.Done()

		// Create a once so the errd channel is only closed once.
		var once sync.Once

		for {
			err, ok := <-f.errC
			if !ok { // channel closed so we can exit.
				return
			}

			attrs := []any{}

			if v := new(RuntimeError); errors.As(err, v) {
				attrs = append(attrs, slog.String("stack", string(v.Stack)))
			}

			if v := new(CleanupError); errors.As(err, v) {
				attrs = append(attrs, slog.String("stack", string(v.Stack)))
			}

			// Log the error.
			slog.Error(err.Error(), attrs...)

			// Close the errd channel. This will cause the below go routine to unblock on the select and thus call Stop().
			// It will also set the os.Exit code to a non zero value indicating an error during execution.
			once.Do(func() {
				exitCode = 1
				close(errd)
			})
		}
	}()

	// Start a go routine which waits for an OS signal, an error is encountered, or all functions exit.
	// Will always call Stop() so clean up functions are called.
	go func() {
		defer wg.Done()

		// Channel to receive os signals on.
		ch := make(chan os.Signal, 1)

		// Notify onto the channel SIGINT, SIGTERM, SIGQUIT events
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		select {
		case <-done:
			// All functions exited normally so we do not need to wait so we can exit out.
		case <-errd:
			// An error occurred during runtime so we should stop.
		case sig := <-ch:
			// Received an os signal to explicitly exit.
			slog.Debug("received os signal", slog.String("signal", sig.String()))
		}

		// Stop listening for OS Signals
		signal.Stop(ch)

		// Stop anything that's running.
		slog.Debug("stop foundation")
		f.stop()
	}()

	// Run the given runner.
	f.Run(ctx, runner)

	// Wait for function to complete.
	<-f.wait()

	// Close the done channel.
	close(done)

	// Wait for go routines to exit
	wg.Wait()

	// Execute finalisers and log their errors.
	for _, fn := range f.finalisers {
		if err := fn(); err != nil {
			exitCode = 1
			slog.Error(err.Error())
		}
	}

	// Call os.Exit once everything is done, if we erroed this will be a none zero exit code.
	os.Exit(exitCode)
}
