package probe

import "context"

// A Sensor is a health check probe sensor which determines if an something
// is healthy.
type Sensor interface {
	Name() string
	Mode() Mode
	Run(ctx context.Context) error
}

// A SensorFunc is a functiontion called by a sensor to determine the health of the sensor.
type SensorFunc func(ctx context.Context) error

// NewSensor constructs a new Sensor.
func NewSensor(name string, mode Mode, f SensorFunc) Sensor {
	return &sensor{
		name: name,
		mode: mode,
		f:    f,
	}
}

type sensor struct {
	name string
	mode Mode
	f    SensorFunc
}

func (s *sensor) Name() string                  { return s.name }
func (s *sensor) Mode() Mode                    { return s.mode }
func (s *sensor) Run(ctx context.Context) error { return s.f(ctx) }
