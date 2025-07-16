package health

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"go.krak3n.io/foundation/health/probe"
)

// ServeMux returns a *http.ServeMux for routing http requests to the HTTP health check handler.
func ServeMux(prefix string, handler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	if prefix != "" {
		mux.Handle(fmt.Sprintf("GET %s", prefix), handler)
	}

	mux.Handle(fmt.Sprintf("GET %s/{$}", prefix), handler)
	mux.Handle(fmt.Sprintf("GET %s/{mode}", prefix), handler)

	return mux
}

// A Handler is a HTTP handler for serving the HTTP health check endpoint.
type Handler struct {
	registry  SensorRegistry
	marshaler ReportsMarshaler
}

// JSONHandler returns a JSON HTTP health check endpoint handler.
func JSONHandler() http.Handler {
	return &Handler{
		registry:  DefaultSensorRegistry(),
		marshaler: JSONReportMarshaler(),
	}
}

// ServeHTTP runs the sensors capturing the status and writing the report back on the response.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	mode := probe.AllModes

	if v := r.PathValue("mode"); v != "" {
		var ok bool

		if mode, ok = probe.ModeFromString(v); !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}
	}

	sensors := slices.DeleteFunc(slices.Clone(h.registry.Sensors()), func(s probe.Sensor) bool {
		return s.Mode()&mode == 0
	})

	status := http.StatusOK

	reports := make([]Report, 0)

	for s := range probe.Run(ctx, sensors...) {
		if s.Status == probe.StatusFailed {
			status = http.StatusServiceUnavailable
		}

		reports = append(reports, Report{
			Name:   s.Name,
			Mode:   s.Mode,
			Status: s.Status,
		})
	}

	b, err := h.marshaler.MarshalReports(reports...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "failed to marshal health probe sensor reports", slog.String("err", err.Error()))

		return
	}

	w.Header().Set("Content-Type", h.marshaler.ContentType())
	w.WriteHeader(status)

	if _, err := w.Write(b); err != nil {
		slog.ErrorContext(ctx, "failed to write health probe sensor reports", slog.String("err", err.Error()))
	}
}
