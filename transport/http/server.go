package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health/probe"
)

func Run(handler http.Handler) foundation.Runner {
	return foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Parallel()

		mux := http.NewServeMux()
		mux.Handle("GET /", handler)
		mux.Handle("GET /_probe", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// TODO:: Configurable options
		server := &http.Server{
			Addr:    "localhost:3000",
			Handler: mux,
		}

		f.On().Stop(func() {
			if err := server.Shutdown(ctx); err != nil {
				f.Error(err)
			}
		})

		url := url.URL{
			Scheme: "http", // TODO: configurable according to the servers TLS config
			Host:   server.Addr,
			Path:   "/_probe",
		}

		probe.Register(Sensor(url.String()))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			f.Error(err)
		}
	})
}
