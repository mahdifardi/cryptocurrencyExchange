package orderbook

import (
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
)

func (ob *Orderbook) PlaceLimitOrder(price float64, o *limit.LimitOrder) {
	var limitOfOrder *limit.Limit

	ob.mu.Lock()
	defer ob.mu.Unlock()

	if o.Bid {
		limitOfOrder = ob.BidLimits[price]
	} else {
		limitOfOrder = ob.AskLimits[price]
	}

	if limitOfOrder == nil {
		limitOfOrder = limit.NewLimit(price)
		limitOfOrder.AddOrder(o)
		if o.Bid {
			ob.bids = append(ob.bids, limitOfOrder)
			ob.BidLimits[price] = limitOfOrder
		} else {
			ob.asks = append(ob.asks, limitOfOrder)
			ob.AskLimits[price] = limitOfOrder
		}
	} else {
		limitOfOrder.AddOrder(o)

	}
	ob.Orders[o.ID] = o

}
