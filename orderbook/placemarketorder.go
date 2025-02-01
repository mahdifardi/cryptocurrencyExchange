package orderbook

import (
	"fmt"
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ob *Orderbook) PlaceMarketOrder(o *limit.LimitOrder, market order.Market) []limit.Match {
	matches := []limit.Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("market [%s] | not enough volume [size: %.2f] for market order [size: %.2f]", market, ob.AskTotalVolume(), o.Size))
		}
		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit)
			}

		}
	} else {
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("market [%s] | not enough volume [size: %.2f] for market order [size: %.2f]", market, ob.BidTotalVolume(), o.Size))
		}
		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}

		}

	}

	for _, match := range matches {
		trade := &Trade{
			Price:     match.Price,
			Size:      match.SizeFilled,
			Bid:       o.Bid,
			Timestamp: time.Now().UnixNano(),
		}

		ob.Trades = append(ob.Trades, trade)

	}

	return matches
}
