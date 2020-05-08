# orderbook

The `orderbook` package is a simple limit order book implemented as a go exercise. Currently limit buys, sells, and order cancellation are all supported.  

## Example

```go
package main

import (
    bx "github.com/fogonthedowns/orderbook"
)

func main() {
    Actions := make(chan *bx.Action)
    done := make(chan bool)

    go bx.ConsoleActionHandler(Actions, done)

    ob := bx.NewOrderBook(Actions)
    ob.AddOrder(bx.NewOrder(1, false, 50, 50))
    ob.AddOrder(bx.NewOrder(2, false, 45, 25))
    ob.AddOrder(bx.NewOrder(3, false, 45, 25))
    ob.AddOrder(bx.NewOrder(4, true, 55, 75))
    ob.CancelOrder(1)
    ob.Done()

    <-done
}
```

As the order book receives commands it generates action messages that needed to be handled to ensure durability. Two channels are used in this example: 

- The `Actions` channel is written to by the order book and read by an action handler. For debugging purposes a `ConsoleActionHandler` is included that simply prints Actions to the console as they arrive.
- When the `Done` command is issued the program should block on the action handler to finish processing outstanding Actions. The `done` channel allows this syncrhonization.

## Future features

- Serialize orderbook reads and writes to add consistency across trader clients
- Add a websocket layer and web interface 
