package orderbook

import (
	"sort"

	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ob *Orderbook) Asks() []*limit.Limit {
	sort.Sort(limit.ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *Orderbook) Bids() []*limit.Limit {
	sort.Sort(limit.ByBestBid{ob.bids})
	return ob.bids
}

func (ob *Orderbook) BidTotalVolume() float64 {
	bidTotalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		bidTotalVolume += ob.bids[i].TotalVolume
	}

	return bidTotalVolume
}

func (ob *Orderbook) AskTotalVolume() float64 {
	askTotalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		askTotalVolume += ob.asks[i].TotalVolume
	}

	return askTotalVolume
}

func (ob *Orderbook) StopLimits() []*order.StopOrder {
	sort.Sort(order.StopOrders(ob.stopLimitOrders))
	return ob.stopLimitOrders
}

func (ob *Orderbook) StopMarkets() []*order.StopOrder {
	sort.Sort(order.StopOrders(ob.stopMarketOrders))
	return ob.stopMarketOrders
}
