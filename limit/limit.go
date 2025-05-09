package limit

import (
	"fmt"
)

type Limit struct {
	Price       float64
	Orders      LimitOrders
	TotalVolume float64
}

type Limits []*Limit

func (l Limits) Len() int      { return len(l) }
func (l Limits) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

type ByBestAsk struct{ Limits }

// func (a ByBestAsk) Len() int           { return len(a.Limits) }
// func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

// func (b ByBestBid) Len() int           { return len(b.Limits) }
// func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*LimitOrder{},
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("[price: %.2f | volume: %.2f]", l.Price, l.TotalVolume)
}

func (l *Limit) AddOrder(o *LimitOrder) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}
