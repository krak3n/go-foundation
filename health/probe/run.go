package probe

import (
	"context"
	"slices"
	"sync"
)

// A SensorStatus is the status of a Sensor.
type SensorStatus struct {
	Name   string
	Mode   Mode
	Status Status
}

// Run executes the given sensors in go routines returning a channel of sensor reports describing
// the result of the sensor.
func Run(ctx context.Context, sensors ...Sensor) <-chan SensorStatus {
	ch := make(chan SensorStatus)

	go func() {
		defer close(ch)

		var wg sync.WaitGroup
		wg.Add(len(sensors))

		for sensor := range slices.Values(sensors) {
			go func(sensor Sensor) {
				defer wg.Done()

				if sensor == nil {
					return
				}

				status := StatusSuccess

				if err := sensor.Run(ctx); err != nil {
					status = StatusFailed
				}

				ch <- SensorStatus{
					Name:   sensor.Name(),
					Mode:   sensor.Mode(),
					Status: status,
				}
			}(sensor)
		}

		wg.Wait()
	}()

	return ch
}
