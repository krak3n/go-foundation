package health

import (
	"log/slog"

	"go.krak3n.io/foundation/health/probe"
)

// A Report is a probe sensor status report.
type Report struct {
	Name   string       `json:"name"`
	Mode   probe.Mode   `json:"mode"`
	Status probe.Status `json:"status"`
}

// A ReportsMarshaler can marshal Report's for the HTTP server.
type ReportsMarshaler interface {
	slog.LogValuer

	MarshalReports(reports ...Report) ([]byte, error)
	ContentType() string
}
