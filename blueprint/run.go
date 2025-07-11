package blueprint

import (
	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health"
)

// Run runs the given runner with in a standard opinionated set of other runners which provides
// telemetry, logging, healthchecks etc.
func Run(name string, r foundation.Runner) {
	foundation.Run(name, health.Run(r))
}
