package main

import (
	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health"
	"go.krak3n.io/foundation/transport/http"
)

func main() {
	foundation.Run("http", health.Run(http.Run(Handler())))
}
