package exchange

import (
	"fmt"
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

var (
	tick = 1 * time.Second
)

func (ex *Exchange) ProcessStopLimitOrders(market order.Market) {
	ticker := time.NewTicker(tick)

	for {

		ob := ex.Orderbook[market]

		// simple search, but becuse ob.StopLimits() is sorted, it should refactored to binary search
		for _, stopLimitOrder := range ob.StopLimits() {
			exchangePrice := ob.Trades[len(ob.Trades)-1].Price

			if stopLimitOrder.State == order.Triggered {
				continue
			}

			shouldTrigger := false
			if stopLimitOrder.Bid && stopLimitOrder.StopPrice >= exchangePrice {
				shouldTrigger = true
			} else if !stopLimitOrder.Bid && stopLimitOrder.StopPrice <= exchangePrice {
				shouldTrigger = true
			}

			if shouldTrigger {
				limitOrder := limit.NewLimitOrder(stopLimitOrder.Bid, stopLimitOrder.Size, stopLimitOrder.UserID)
				ob.PlaceLimitOrder(stopLimitOrder.Price, limitOrder)
				stopLimitOrder.State = order.Triggered

				if stopLimitOrder.Bid {

					fmt.Printf("stop Bid limit order triggered =>%d | price [%.2f] | size [%.2f]", stopLimitOrder.ID, stopLimitOrder.Price, stopLimitOrder.Size)
				} else {
					fmt.Printf("stop Ask limit order triggered =>%d | price [%.2f] | size [%.2f]", stopLimitOrder.ID, stopLimitOrder.Price, stopLimitOrder.Size)

				}
			}

		}
		<-ticker.C
	}

}

func (ex *Exchange) ProcessStopMarketOrders(market order.Market) {
	ticker := time.NewTicker(tick)

	for {

		ob := ex.Orderbook[market]

		// simple search, but becuse ob.StopMarkets() is sorted, it should refactored to binary search
		for _, stopMarketOrder := range ob.StopMarkets() {
			exchangePrice := ob.Trades[len(ob.Trades)-1].Price

			if stopMarketOrder.State == order.Triggered {
				continue
			}

			shouldTrigger := false
			if stopMarketOrder.Bid && stopMarketOrder.StopPrice >= exchangePrice {
				shouldTrigger = true
			} else if !stopMarketOrder.Bid && stopMarketOrder.StopPrice <= exchangePrice {
				shouldTrigger = true
			}

			if shouldTrigger {
				marketOrder := limit.NewLimitOrder(stopMarketOrder.Bid, stopMarketOrder.Size, stopMarketOrder.UserID)
				ob.PlaceMarketOrder(marketOrder, market)
				stopMarketOrder.State = order.Triggered

				if stopMarketOrder.Bid {

					fmt.Printf("stop Bid Market order triggered =>%d | price [%.2f] | size [%.2f]", stopMarketOrder.ID, stopMarketOrder.Price, stopMarketOrder.Size)
				} else {
					fmt.Printf("stop Ask Market order triggered =>%d | price [%.2f] | size [%.2f]", stopMarketOrder.ID, stopMarketOrder.Price, stopMarketOrder.Size)

				}
			}

		}
		<-ticker.C
	}

}
