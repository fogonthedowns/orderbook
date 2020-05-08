package orderbook

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestBehavior(t *testing.T) {
	Actions := make(chan *Action)
	done := make(chan bool)
	ob := NewOrderBook(Actions)

	log := make([]*Action, 0)
	go func() {
		for {
			action := <-Actions
			log = append(log, action)
			if action.ActionType == AT_DONE {
				done <- true
				return
			}
		}
	}()

	// Should all go into the book
	ob.AddOrder(&Order{IsBuy: false, Id: "1", Price: 50, Amount: 50})
	ob.AddOrder(&Order{IsBuy: false, Id: "2", Price: 45, Amount: 25})
	ob.AddOrder(&Order{IsBuy: false, Id: "3", Price: 45, Amount: 25})
	// Should trigger three fills, two partial at 45 and one at 50
	ob.AddOrder(&Order{IsBuy: true, Id: "4", Price: 55, Amount: 75})
	// Should cancel immediately
	ob.CancelOrder("1")
	// Should all go into the book
	ob.AddOrder(&Order{IsBuy: true, Id: "5", Price: 55, Amount: 20})
	ob.AddOrder(&Order{IsBuy: true, Id: "6", Price: 50, Amount: 15})
	// Should trigger two fills, one partial at 55 and one at 50
	ob.AddOrder(&Order{IsBuy: false, Id: "7", Price: 45, Amount: 25})
	ob.Done()

	<-done

	expected := []*Action{
		&Action{AT_SELL, "1", "", 50, 50},
		&Action{AT_SELL, "2", "", 25, 45},
		&Action{AT_SELL, "3", "", 25, 45},
		&Action{AT_BUY, "4", "", 75, 55},
		&Action{AT_PARTIAL_FILLED, "4", "2", 25, 45},
		&Action{AT_PARTIAL_FILLED, "4", "3", 25, 45},
		&Action{AT_FILLED, "4", "1", 25, 50},
		&Action{AT_CANCEL, "1", "", 0, 0},
		&Action{AT_CANCELLED, "1", "", 0, 0},
		&Action{AT_BUY, "5", "", 20, 55},
		&Action{AT_BUY, "6", "", 15, 50},
		&Action{AT_SELL, "7", "", 25, 45},
		&Action{AT_PARTIAL_FILLED, "7", "5", 20, 55},
		&Action{AT_FILLED, "7", "6", 5, 50},
		&Action{AT_DONE, "", "", 0, 0},
	}
	if !reflect.DeepEqual(log, expected) {
		t.Error("\n\nExpected:\n\n", expected, "\n\nGot:\n\n", log, "\n\n")
	}
}

func buildOrders(n int, PriceMean, PriceStd float64, maxAmount int32) []*Order {
	orders := make([]*Order, 0)
	var Price uint32
	for i := 0; i < n; i++ {
		Price = uint32(math.Abs(rand.NormFloat64()*PriceStd + PriceMean))
		orders = append(orders, &Order{
			Id:     strconv.Itoa(i + 1),
			IsBuy:  float64(Price) >= PriceMean,
			Price:  Price,
			Amount: uint32(rand.Int31n(maxAmount)),
		})
	}
	return orders
}

func doPerfTest(n int, PriceMean, PriceStd float64, maxAmount int32) {
	orders := buildOrders(n, PriceMean, PriceStd, maxAmount)
	Actions := make(chan *Action)
	done := make(chan bool)
	ob := NewOrderBook(Actions)
	actionCount := 0

	go func() {
		for {
			action := <-Actions
			actionCount++
			if action.ActionType == AT_DONE {
				done <- true
				return
			}
		}
	}()

	start := time.Now()
	for _, order := range orders {
		ob.AddOrder(order)
	}
	ob.Done()
	<-done
	elapsed := time.Since(start)

	fmt.Printf("Handled %v Actions in %v at %v Actions/second.\n",
		actionCount, elapsed, int(float64(actionCount)/elapsed.Seconds()))
}

func TestPerf(t *testing.T) {
	doPerfTest(10000, 5000, 10, 50)
	doPerfTest(10000, 5000, 1000, 5000)
	doPerfTest(100000, 5000, 10, 50)
	doPerfTest(100000, 5000, 1000, 5000)
	doPerfTest(1000000, 5000, 10, 50)
	doPerfTest(1000000, 5000, 1000, 5000)
}
