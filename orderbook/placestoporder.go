package orderbook

import "github.com/mahdifardi/cryptocurrencyExchange/order"

func (ob *Orderbook) PlaceStopOrder(o *order.StopOrder) {
	if o.Limit {
		ob.stopLimitOrders = append(ob.stopLimitOrders, o)
	} else {
		ob.stopMarketOrders = append(ob.stopMarketOrders, o)
	}
}
