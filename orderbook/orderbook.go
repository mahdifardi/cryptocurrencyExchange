package orderbook

import (
	"sync"

	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

type Trade struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type Orderbook struct {
	asks []*limit.Limit
	bids []*limit.Limit

	Trades []*Trade
	mu     sync.RWMutex

	AskLimits map[float64]*limit.Limit
	BidLimits map[float64]*limit.Limit

	Orders map[int64]*limit.LimitOrder

	stopLimitOrders  []*order.StopOrder
	stopMarketOrders []*order.StopOrder
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:   []*limit.Limit{},
		bids:   []*limit.Limit{},
		Trades: []*Trade{},

		AskLimits: make(map[float64]*limit.Limit),
		BidLimits: make(map[float64]*limit.Limit),
		Orders:    make(map[int64]*limit.LimitOrder),

		stopLimitOrders:  []*order.StopOrder{},
		stopMarketOrders: []*order.StopOrder{},
	}
}
