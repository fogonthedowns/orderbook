package orderbook

import "fmt"

type ActionType string

const (
	AT_BUY            = "BUY"
	AT_SELL           = "SELL"
	AT_CANCEL         = "CANCEL"
	AT_CANCELLED      = "CANCELLED"
	AT_PARTIAL_FILLED = "PARTIAL_FILLED"
	AT_FILLED         = "FILLED"
	AT_DONE           = "DONE"
)

type Action struct {
	ActionType  ActionType `json:"ActionType"`
	OrderId     string     `json:"OrderId"`
	FromOrderId string     `json:"FromOrderId"`
	Amount      uint32     `json:"Amount"`
	Price       uint32     `json:"Price"`
}

func (a *Action) String() string {
	return fmt.Sprintf("\nAction{ActionType:%v,OrderId:%v,FromOrderId:%v,Amount:%v,Price:%v}",
		a.ActionType, a.OrderId, a.FromOrderId, a.Amount, a.Price)
}

func NewBuyAction(o *Order) *Action {
	return &Action{ActionType: AT_BUY, OrderId: o.Id, Amount: o.Amount,
		Price: o.Price}
}

func NewSellAction(o *Order) *Action {
	return &Action{ActionType: AT_SELL, OrderId: o.Id, Amount: o.Amount,
		Price: o.Price}
}

func NewCancelAction(Id string) *Action {
	return &Action{ActionType: AT_CANCEL, OrderId: Id}
}

func NewCancelledAction(Id string) *Action {
	return &Action{ActionType: AT_CANCELLED, OrderId: Id}
}

func NewPartialFilledAction(o *Order, fromOrder *Order) *Action {
	return &Action{ActionType: AT_PARTIAL_FILLED, OrderId: o.Id, FromOrderId: fromOrder.Id,
		Amount: fromOrder.Amount, Price: fromOrder.Price}
}

func NewFilledAction(o *Order, fromOrder *Order) *Action {
	return &Action{ActionType: AT_FILLED, OrderId: o.Id, FromOrderId: fromOrder.Id,
		Amount: o.Amount, Price: fromOrder.Price}
}

func NewDoneAction() *Action {
	return &Action{ActionType: AT_DONE}
}

func ConsoleActionHandler(Actions <-chan *Action, done chan<- bool) {
	for {
		a := <-Actions
		switch a.ActionType {
		case AT_BUY, AT_SELL:
			fmt.Printf("%s - Order: %v, Amount: %v, Price: %v\n",
				a.ActionType, a.OrderId, a.Amount, a.Price)
		case AT_CANCEL, AT_CANCELLED:
			fmt.Printf("%s - Order: %v\n", a.ActionType, a.OrderId)
		case AT_PARTIAL_FILLED, AT_FILLED:
			fmt.Printf("%s - Order: %v, Filled %v@%v, From: %v\n",
				a.ActionType, a.OrderId, a.Amount, a.Price, a.FromOrderId)
		case AT_DONE:
			fmt.Printf("%s\n", a.ActionType)
			done <- true
			return
		default:
			panic("Unknown action type.")
		}
	}
}

func NoopActionHandler(Actions <-chan *Action) {
	for {
		<-Actions
	}
}
