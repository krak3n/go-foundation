package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health/probe"
)

type RunnerOption interface {
	applyHTTPServer(*http.Server)
}

type RunnerOptions []RunnerOption

func (o RunnerOptions) applyHTTPServer(srv *http.Server) {
	for opt := range slices.Values(o) {
		if opt != nil {
			opt.applyHTTPServer(srv)
		}
	}
}

type RunnerOptionFunc func(*http.Server)

func (f RunnerOptionFunc) applyHTTPServer(srv *http.Server) {
	f(srv)
}

func WtihServerAddress(addr string) RunnerOption {
	return RunnerOptionFunc(func(s *http.Server) {
		s.Addr = addr
	})
}

func Run(handler http.Handler, opts ...RunnerOption) foundation.Runner {
	return foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		mux := http.NewServeMux()
		mux.Handle("GET /", handler)
		mux.Handle("GET /_sensor", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		server := &http.Server{
			Addr:    "127.0.0.1:3000",
			Handler: mux,
		}

		RunnerOptions(opts).applyHTTPServer(server)

		f.On().Stop(func() {
			if err := server.Shutdown(ctx); err != nil {
				f.Error(err)
			}
		})

		url := url.URL{
			Scheme: "http", // TODO: configurable according to the servers TLS config
			Host:   server.Addr,
			Path:   "/_sensor",
		}

		probe.Register(Sensor(url.String()))

		f.Parallel() // Mark the Runner as parallel now we are going start blocking

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			f.Error(err)
		}
	})
}
