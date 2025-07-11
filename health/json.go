package health

import (
	"encoding/json"
	"log/slog"
)

func JSONReportMarshaler() ReportsMarshaler {
	return &jsonReportMarshaler{
		marshaler: json.Marshal,
	}
}

type jsonReportMarshaler struct {
	marshaler func(v any) ([]byte, error)
}

func (m *jsonReportMarshaler) LogValue() slog.Value {
	return slog.StringValue("JSON")
}

func (m *jsonReportMarshaler) ContentType() string {
	return "application/json"
}

func (m *jsonReportMarshaler) MarshalReports(reports ...Report) ([]byte, error) {
	return m.marshaler(reports)
}
