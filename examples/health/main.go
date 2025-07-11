package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/health"
	"go.krak3n.io/foundation/health/probe"
)

func main() {
	runner := foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.On().Done(func() {
				fmt.Println("done", f.Name())
			})

			f.On().Stop(func() {
				fmt.Println("stop", f.Name())
			})

			f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
				f.Parallel()

				probe.Register(probe.NewSensor("sensor1", probe.AllModes, func(context.Context) error {
					return nil
				}))

				c := make(chan struct{})

				f.On().Stop(func() {
					fmt.Println("close c", f.Name())
					close(c)
				})

				f.On().Done(func() {
					fmt.Println("done c", f.Name())
				})

				fmt.Println("block on c", f.Name())
				<-c
				fmt.Println("c unblocked", f.Name())
			}))
		}))

		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.Parallel()

			probe.Register(probe.NewSensor("sensor2", probe.StartupLivenessMode, func(context.Context) error {
				return nil
			}))

			c := make(chan struct{})

			f.On().Stop(func() {
				fmt.Println("close c", f.Name())
				close(c)
			})

			f.On().Done(func() {
				fmt.Println("done c", f.Name())
			})

			fmt.Println("block on c", f.Name())
			<-c
			fmt.Println("c unblocked", f.Name())
		}))
	})

	foundation.Run("health", health.Run(runner))
}
