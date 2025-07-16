package http

import (
	"context"
	"fmt"
	"net/http"

	"go.krak3n.io/foundation/health/probe"
)

// Sensor returns a health probe sensor for HTTP servers.
// The sensor makes a HTTP GET request to the given url, the response must be a 200 OK for the sensor
// to return a healthy status.
func Sensor(url string) probe.Sensor {
	client := http.DefaultClient

	return probe.NewSensor("http.server", probe.AllModes, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("construct http request: %w", err)
		}

		rsp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("make client request: %w", err)
		}

		if err := rsp.Body.Close(); err != nil {
			return fmt.Errorf("close response body: %w", err)
		}

		if code := rsp.StatusCode; code != http.StatusOK {
			return fmt.Errorf("invalid status code %d", code)
		}

		return nil
	})
}
