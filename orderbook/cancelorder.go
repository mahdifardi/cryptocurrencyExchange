package orderbook

import (
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ob *Orderbook) CancelOrder(o *limit.LimitOrder) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)

	if len(limit.Orders) == 0 {
		ob.clearLimit(o.Bid, limit)
	}
}

func (ob *Orderbook) CancelStopOrder(so *order.StopOrder) {
	if so.Limit {
		for i := 0; i < len(ob.stopLimitOrders); i++ {
			if ob.stopLimitOrders[i] == so && ob.stopLimitOrders[i].State == order.Pending {
				ob.stopLimitOrders[i].State = order.Canceled
				// ob.stopLimitOrders[i] = ob.stopLimitOrders[len(ob.stopLimitOrders)-1]
				// ob.stopLimitOrders = ob.stopLimitOrders[:len(ob.stopLimitOrders)-1]

			}
		}
	} else {
		for i := 0; i < len(ob.stopMarketOrders); i++ {
			if ob.stopMarketOrders[i] == so && ob.stopMarketOrders[i].State == order.Pending {
				ob.stopMarketOrders[i].State = order.Canceled
				// ob.stopMarketOrders[i] = ob.stopMarketOrders[len(ob.stopMarketOrders)-1]
				// ob.stopMarketOrders = ob.stopMarketOrders[:len(ob.stopMarketOrders)-1]
			}
		}
	}
}
