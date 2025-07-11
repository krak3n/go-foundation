package probe

import (
	"sync"
)

var globalRegistry = &registry{sensors: make([]Sensor, 0)}

// Register registers one or more sensors.
func Register(sensors ...Sensor) {
	globalRegistry.Register(sensors...)
}

// Sensors returns the registered sensors.
func Sensors() []Sensor {
	return globalRegistry.Sensors()
}

type registry struct {
	mtx     sync.RWMutex
	sensors []Sensor
}

// Register registers a sensor.
func (r *registry) Register(sensors ...Sensor) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.sensors = append(r.sensors, sensors...)
}

// Sensors returns the sensors filtered by mode.
func (r *registry) Sensors() []Sensor {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.sensors
}
