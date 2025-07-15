# Foundation

<p align="center">
  <img src="logo.png" alt="Foundation Gopher Mascot" width="300"/>
</p>

Foundation is a simple Go Application framework inspired by Go's `testing` framework. It does not make any assumptions about what and how you want things to run but it does provide the building blocks to run your application your way.

However if you would like an opinionated way of running your application which will come with batteries included components like OpenTelemetry and health check server support you should checkout the `blueprint` package.

## Core Concepts

Foundation is inspired by Go's `testing` framework. This essentially Foundation will run a something, usually a function but can also be a more complex type if it implements the `Runner` interface. This `Runner` can be a blocking or non blocking, these `Runner`s can also spawn other `Runner`s which again can be blocking or non blocking. Foundation will ensure that all this `Runner`s are executed in order and stopped in order.

Below is a simple example:

```go
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
```

The output from this would be:

```
Do Some Work in: simple.1
Do Some Work in: simple.1.1
Do Some Work in: simple.1.2
Done Some Work in: simple.1.2
Done Some Work in: simple.1.1
Done Some Work in: simple.1
```

As you can see everything executes and stops in the order they are declared in code. You will also see the use of `On().Stop()` which we will talk about later.

### Why is `Runner` an `interface`?

This gives the most flexibility to what a `Runner` can do. Sometimes you may just need to run a simple function or some other time you need to run something more complex that needs to store state. Take a look at the `ticker` runner for an example of a more complex `Runner` type.

### Concurrency

As mentioned a `Runner` can be blocking or non blocking. To achieve non blocking behaviour we have implemented the same pattern as the `testing` package with `t.Parallel()`, you can use `f.Parallel()` which will indicate the `Runner` should not block and allow the next `Runner` to run (if there is one).

Here is a simple example:

```go
package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
)

func main() {
	foundation.Run("parallel", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.Parallel()

			c := make(chan struct{})

			f.On().Stop(func() {
				close(c)
			})

			fmt.Println("block on c", f.Name())
			<-c
			fmt.Println("c unblocked", f.Name())
		}))

		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			fmt.Println("after non blocking Runner")
		}))
	}))
}
```

This will give the following output (after a `SIGINT`):

```
block on c parallel.1.1
after non blocking Runner
unblocked parallel.1.1
```

You will see the use of `f.On().Stop()` which is the next feature we will talk about.

### Cleanup

Like the `testing` package we need a way to cleanup when we stop the application but this is slightly different in `foundation` than in `testing`. Because tests are not long lived `testing` can get away with just one `t.Cleanup()` method but `foundation` needs a couple of different concepts.

#### `On().Stop()`

The `On().Stop()` method takes a `func()`, you can call `Stop()` multiple times to register multiple stop functions and like `t.Cleanup` they are called like they would if you used `defer`, last in first out. Stop functions are called either when the application exists normally without any blocking `Runner`s or when there is an explicit signal to stop the application, for example a os `SIGINT` signal.

```go
package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
)

func main() {
	foundation.Run("parallel", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.Parallel()

			c := make(chan struct{})

			f.On().Stop(func() {
				fmt.Println("stop func 2")
				close(c)
			})

			f.On().Stop(func() {
				fmt.Println("stop func 1")
			})

			fmt.Println("block on c", f.Name())
			<-c
			fmt.Println("c unblocked", f.Name())
		}))
	}))
}
```

This will produce the following output after a `SIGINT`.

```
block on c parallel.1.1
stop func 1
stop func 2
c unblocked parallel.1.1
```

#### `On().Done()`

Functions registered using `On().Done()` will always be called when `Runner` has completed, these happen **after** `Stop()` functions stop are similar to `t.Cleanup()`. Like `Stop()` functions are called last in first out order.

```go
package main

import (
	"context"
	"fmt"

	"go.krak3n.io/foundation"
)

func main() {
	foundation.Run("parallel", foundation.RunFunc(func(ctx context.Context, f foundation.F) {
		f.Run(ctx, foundation.RunFunc(func(ctx context.Context, f foundation.F) {
			f.Parallel()

			c := make(chan struct{})

			f.On().Stop(func() {
				fmt.Println("stop func 2")
				close(c)
			})

			f.On().Stop(func() {
				fmt.Println("stop func 1")
			})

			f.On().Done(func() {
				fmt.Println("done func 2")
			})

			f.On().Done(func() {
				fmt.Println("done func 1")
			})

			fmt.Println("block on c", f.Name())
			<-c
			fmt.Println("c unblocked", f.Name())
		}))
	}))
}
```

This will produce the following output after a `SIGINT`.

```
block on c parallel.1.1
stop func 1
stop func 2
c unblocked parallel.1.1
done func 1
done func 2
```
