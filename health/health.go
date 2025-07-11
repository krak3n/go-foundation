package health

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health/probe"
)

func Run(runners ...foundation.Runner) foundation.Runner {
	return NewRunner(runners...)
}

type Runner struct {
	runners []foundation.Runner
}

func NewRunner(runners ...foundation.Runner) *Runner {
	return &Runner{
		runners: runners,
	}
}

func (r *Runner) Run(ctx context.Context, f foundation.F) {
	var wg sync.WaitGroup

	wg.Add(1)

	server := &http.Server{
		Addr: "127.0.0.1:3417",
	}

	var serving bool

	f.On().Stop(func() {
		if serving {
			if err := server.Shutdown(ctx); err != nil {
				f.Error(err)
			}
		}

		wg.Wait()
	})

	f.Run(ctx, r.runners...)

	server.Handler = ServeMux("/_health", JSONHandler(probe.Sensors()...))

	go func() {
		defer wg.Done()

		serving = true
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serving = false

			f.Error(err)
		}
	}()
}
