package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
)

func main() {
	foundation.Run("simple", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.On().Stop(func() {
			fmt.Println("Done Some Work in:", f.Name())
		})

		fmt.Println("Do Some Work in:", f.Name())

		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.On().Stop(func() {
				fmt.Println("Done Some Work in:", f.Name())
			})

			fmt.Println("Do Some Work in:", f.Name())
		}))

		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.On().Stop(func() {
				fmt.Println("Done Some Work in:", f.Name())
			})

			fmt.Println("Do Some Work in:", f.Name())
		}))
	}))
}
