package health

import (
	"context"
	stdhttp "net/http"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/transport/http"
)

// Run returns a foundation.Runner which runs a standard HTTP server on 127.0.0.1:3417.
// The server will only response with a non 503 response until all runners have registered their
// sensors and all sensors do not error.
// As soon as a stop signal is received the server will respond with a 503.
// The server is the last thing to stop.
func Run(runners ...foundation.Runner) foundation.Runner {
	return foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		// Track the state of whether we want the health check server to response available or not.
		// We want the server to the first thing we start but to only allow sensors to be checked
		// once all runners have run and therefore registered their sensors.
		// We also want the server to be the last thing to stop but also be marked unavailable immediately
		// before the runners have been told to stop.
		var available bool

		// Start a standard HTTP server serving on 3417 by default
		f.Run(ctx, http.Run(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			if !available {
				w.WriteHeader(stdhttp.StatusServiceUnavailable)

				return
			}

			ServeMux("/_health", JSONHandler()).ServeHTTP(w, r)
		}), http.WtihServerAddress("127.0.0.1:3417")))

		// Add a new runner that is the first to stop which sets the HTTP health check server as unavailable
		runners := append(runners, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.On().Stop(func() {
				available = false
			})
		}))

		// Now all probes should be registered we can mark the server as generally available
		f.On().Done(func() {
			available = true
		})

		// Run the runners
		f.Run(ctx, runners...)
	})
}
