package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
)

func main() {
	foundation.Run("parallel", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.On().Done(func() {
				fmt.Println("done", f.Name())
			})

			f.On().Stop(func() {
				fmt.Println("stop", f.Name())
			})

			f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
				f.Parallel()

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
}
