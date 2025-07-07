package main

import (
	"context"
	"fmt"
	"time"

	"go.krak3n.io/foundation"
	"go.krak3n.io/foundation/tick"
)

func main() {
	foundation.Run("ticker-example", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		tick.Run(ctx, f, time.Second, func(ctx context.Context, t tick.Ticker) {
			fmt.Println(fmt.Sprintf("ticker %s tick at: %s", t.Name(), t.Tick()))
		})

		tick.Run(ctx, f, time.Second*2, func(ctx context.Context, t tick.Ticker) {
			t.On().Stop(func() {
				fmt.Println(fmt.Sprintf("stop ticker %s tick at: %s", t.Name(), t.Tick()))
			})

			t.On().Done(func() {
				fmt.Println(fmt.Sprintf("done ticker %s tick at: %s", t.Name(), t.Tick()))
			})

			fmt.Println(fmt.Sprintf("ticker %s tick at: %s", t.Name(), t.Tick()))
		}, tick.WithUntil(5))

		tick.Exponential(ctx, f, 10, time.Millisecond*200, func(ctx context.Context, t tick.Ticker) {
			t.On().Done(func() {
				fmt.Println(fmt.Sprintf("done expoential ticker %s tick at: %s", t.Name(), t.Tick()))
			})

			fmt.Println(fmt.Sprintf("expoential ticker %s tick at: %s", t.Name(), t.Tick()))
		})
	}))
}
