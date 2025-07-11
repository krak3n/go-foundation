package probe

import (
	"log/slog"
	"strconv"
)

// Supported probe sensor statuses.
const (
	StatusFailed Status = iota + 1
	StatusSuccess
)

// A Status is returned by a sensor indicating whether the sensor succeeded or failed.
type Status int8

func (s Status) String() string {
	var v string

	switch s {
	case StatusFailed:
		v = "failed"
	case StatusSuccess:
		v = "success"
	default:
		v = "unknown"
	}

	return v
}

func (s Status) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

// MarshalJSON marshals a probe status to a valid JSON string.
func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}
