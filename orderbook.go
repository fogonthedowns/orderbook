package orderbook

import "fmt"

const MAX_Price = 10000000

type PricePoint struct {
	orderHead *Order
	orderTail *Order
}

func (pp *PricePoint) Insert(o *Order) {
	if pp.orderHead == nil {
		pp.orderHead = o
		pp.orderTail = o
	} else {
		pp.orderTail.Next = o
		pp.orderTail = o
	}
}

type OrderStatus int

const (
	OS_NEW OrderStatus = iota
	OS_OPEN
	OS_PARTIAL
	OS_FILLED
	OS_CANCELLED
)

type Order struct {
	Id     string
	IsBuy  bool
	Price  uint32
	Amount uint32
	Status OrderStatus
	Next   *Order
}

func (o *Order) String() string {
	return fmt.Sprintf("\nOrder{Id:%v,IsBuy:%v,Price:%v,Amount:%v}",
		o.Id, o.IsBuy, o.Price, o.Amount)
}

func NewOrder(
	Id string,
	IsBuy bool,
	Price uint32,
	Amount uint32,
) *Order {
	return &Order{Id: Id, IsBuy: IsBuy, Price: Price, Amount: Amount,
		Status: OS_NEW}
}

type OrderBook struct {
	// These are more estimates than reportable values
	Ask        uint32
	Bid        uint32
	OrderIndex map[string]*Order
	Prices     [MAX_Price]*PricePoint
	Actions    chan<- *Action
}

func NewOrderBook(Actions chan<- *Action) *OrderBook {
	ob := new(OrderBook)
	ob.Bid = 0
	ob.Ask = MAX_Price
	for i := range ob.Prices {
		ob.Prices[i] = new(PricePoint)
	}
	ob.Actions = Actions
	ob.OrderIndex = make(map[string]*Order)
	return ob
}

func (ob *OrderBook) AddOrder(o *Order) {
	// Try to fill immediately
	if o.IsBuy {
		ob.Actions <- NewBuyAction(o)
		ob.FillBuy(o)
	} else {
		ob.Actions <- NewSellAction(o)
		ob.FillSell(o)
	}

	// Into the book
	if o.Amount > 0 {
		ob.openOrder(o)
	}
}

func (ob *OrderBook) openOrder(o *Order) {
	pp := ob.Prices[o.Price]
	pp.Insert(o)
	o.Status = OS_OPEN
	if o.IsBuy && o.Price > ob.Bid {
		ob.Bid = o.Price
	} else if !o.IsBuy && o.Price < ob.Ask {
		ob.Ask = o.Price
	}
	ob.OrderIndex[o.Id] = o
}

func (ob *OrderBook) FillBuy(o *Order) {
	for (ob.Ask <= o.Price) && (o.Amount > 0) {
		pp := ob.Prices[ob.Ask]
		OrderHead := pp.orderHead
		for OrderHead != nil {
			ob.fill(o, OrderHead)
			OrderHead = OrderHead.Next
			pp.orderHead = OrderHead
		}
		ob.Ask++
	}
}

func (ob *OrderBook) FillSell(o *Order) {
	for (ob.Bid >= o.Price) && (o.Amount > 0) {
		pp := ob.Prices[ob.Bid]
		OrderHead := pp.orderHead
		for OrderHead != nil {
			ob.fill(o, OrderHead)
			OrderHead = OrderHead.Next
			pp.orderHead = OrderHead
		}
		ob.Bid--
	}
}

func (ob *OrderBook) fill(o, OrderHead *Order) {
	if (OrderHead.Amount >= o.Amount) && (o.Amount > 0) {
		ob.Actions <- NewFilledAction(o, OrderHead)
		OrderHead.Amount -= o.Amount
		o.Amount = 0
		o.Status = OS_FILLED
		return
	} else {
		// Partial fill
		if OrderHead.Amount > 0 && (o.Amount > 0) {
			ob.Actions <- NewPartialFilledAction(o, OrderHead)
			o.Amount -= OrderHead.Amount
			o.Status = OS_PARTIAL
			OrderHead.Amount = 0
		}
	}
}

func (ob *OrderBook) CancelOrder(Id string) {
	ob.Actions <- NewCancelAction(Id)
	if o, ok := ob.OrderIndex[Id]; ok {
		// If this is the last order at a particular Price point
		// we need to update the Bid/Ask...right? Maybe not?
		o.Amount = 0
		o.Status = OS_CANCELLED
	}
	ob.Actions <- NewCancelledAction(Id)
}

func (ob *OrderBook) Done() {
	ob.Actions <- NewDoneAction()
}
