package health

import "go.krak3n.io/foundation/health/probe"

type SensorRegistry interface {
	Sensors() []probe.Sensor
}

type SensorRegistryFunc func() []probe.Sensor

func (f SensorRegistryFunc) Sensors() []probe.Sensor {
	return f()
}

func DefaultSensorRegistry() SensorRegistry {
	return SensorRegistryFunc(func() []probe.Sensor {
		return probe.Sensors()
	})
}
