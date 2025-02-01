package orderbook

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := limit.NewLimit(10_000)
	bidOrderA := limit.NewLimitOrder(true, 8, 0)
	bidOrderB := limit.NewLimitOrder(true, 4, 0)
	bidOrderC := limit.NewLimitOrder(true, 6, 0)

	l.AddOrder(bidOrderA)
	l.AddOrder(bidOrderB)
	l.AddOrder(bidOrderC)

	l.DeleteOrder(bidOrderB)

	fmt.Println(22)

	fmt.Println(l)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := limit.NewLimitOrder(false, 10, 0)
	sellOrderB := limit.NewLimitOrder(false, 5, 0)

	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[sellOrderA.ID], sellOrderA)
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB)
	assert(t, len(ob.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := limit.NewLimitOrder(false, 20, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := limit.NewLimitOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(buyOrder, order.MarketETH)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)

	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, matches[0].Bid.IsFilled(), true)

	fmt.Printf("%+v", matches)
}

func TestPlaceMultiMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := limit.NewLimitOrder(true, 5, 0)
	buyOrderB := limit.NewLimitOrder(true, 8, 0)
	buyOrderC := limit.NewLimitOrder(true, 1, 0)
	buyOrderD := limit.NewLimitOrder(true, 1, 0)

	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)

	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	assert(t, ob.BidTotalVolume(), 15.0)

	sellOrder := limit.NewLimitOrder(false, 10, 0)
	matches := ob.PlaceMarketOrder(sellOrder, order.MarketETH)

	assert(t, ob.BidTotalVolume(), 5.0)
	assert(t, len(matches), 2)
	assert(t, len(ob.bids), 2)

	fmt.Printf("%+v", matches)
}

func TestCancelAskOrder(t *testing.T) {
	ob := NewOrderbook()

	price := 10_000.0

	sellOrder := limit.NewLimitOrder(false, 4, 0)
	ob.PlaceLimitOrder(price, sellOrder)

	assert(t, ob.AskTotalVolume(), 4.0)

	assert(t, len(ob.Orders), 1)
	ob.CancelOrder(sellOrder)
	assert(t, len(ob.Orders), 0)
	_, ok := ob.Orders[sellOrder.ID]
	assert(t, ok, false)

	assert(t, ob.AskTotalVolume(), 0.0)

	_, ok = ob.AskLimits[price]
	assert(t, ok, false)
}

func TestCancelBidOrder(t *testing.T) {
	ob := NewOrderbook()

	price := 10_000.0

	buyOrder := limit.NewLimitOrder(true, 4, 0)
	ob.PlaceLimitOrder(price, buyOrder)

	assert(t, ob.BidTotalVolume(), 4.0)

	assert(t, len(ob.Orders), 1)
	ob.CancelOrder(buyOrder)
	assert(t, len(ob.Orders), 0)
	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)

	assert(t, ob.BidTotalVolume(), 0.0)

	_, ok = ob.BidLimits[price]
	assert(t, ok, false)
}

func TestLastMarketTrades(t *testing.T) {
	ob := NewOrderbook()

	price := 10_000.0

	sellOrder := limit.NewLimitOrder(false, 10, 0)
	ob.PlaceLimitOrder(price, sellOrder)

	marketOrder := limit.NewLimitOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(marketOrder, order.MarketETH)

	assert(t, len(matches), 1)
	assert(t, len(ob.Trades), 1)

	trade := ob.Trades[0]
	assert(t, trade.Price, price)
	assert(t, trade.Bid, marketOrder.Bid)
	assert(t, trade.Size, matches[0].SizeFilled)

}
