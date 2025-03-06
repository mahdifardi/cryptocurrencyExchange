package orderbook

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/mahdifardi/cryptocurrencyExchange/config"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func createUser() (*user.User, *user.User) {
	config, err := config.LoadConfig("../config/config.json")
	fmt.Println(config)
	if err != nil {
		log.Fatalf("config file error: %v", err)
	}
	btcUser3Address := config.BtcUser3Address
	ethUser3PrivKey := config.EthUser3Address
	user3 := user.NewUser(ethUser3PrivKey, btcUser3Address, config.User3ID)

	btcUser2Address := config.BtcUser2Address
	ethUser2PrivKey := config.EthUser2Address
	user2 := user.NewUser(ethUser2PrivKey, btcUser2Address, config.User2ID)

	return user3, user2

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
	matches := ob.PlaceMarketOrder(buyOrder, order.MarketETH_Fiat)

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
	matches := ob.PlaceMarketOrder(sellOrder, order.MarketETH_Fiat)

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
	matches := ob.PlaceMarketOrder(marketOrder, order.MarketETH_Fiat)

	assert(t, len(matches), 1)
	assert(t, len(ob.Trades), 1)

	trade := ob.Trades[0]
	assert(t, trade.Price, price)
	assert(t, trade.Bid, marketOrder.Bid)
	assert(t, trade.Size, matches[0].SizeFilled)

}

func TestMakeStopLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	user3, user2 := createUser()

	price := 40_000.0
	sellOrder := limit.NewLimitOrder(true, 3, user2.ID)
	ob.PlaceLimitOrder(price, sellOrder)

	market := order.MarketUSDT_Fiat

	buyMarketOrder := limit.NewLimitOrder(false, 3, user3.ID)
	matches := ob.PlaceMarketOrder(buyMarketOrder, market)

	assert(t, len(matches), 1)
	assert(t, len(ob.Trades), 1)

	assert(t, ob.Trades[len(ob.Trades)-1].Price, price)

	stopBidLimitOrderprice := 38_000.0
	stopBidLimitOrderstopPrice := 42_000.0
	bidStopLimitOrder := order.NewStopOrder(true, true, 4.0, stopBidLimitOrderprice, stopBidLimitOrderstopPrice, 2, market)
	ob.PlaceStopOrder(bidStopLimitOrder, market, user2)

	assert(t, len(ob.stopLimitOrders), 1)

	stopAskLimitOrderprice := 39_000.0
	stopAskLimitOrderstopPrice := 39_000.0
	askStopLimitOrder := order.NewStopOrder(false, true, 9.0, stopAskLimitOrderprice, stopAskLimitOrderstopPrice, 2, market)
	ob.PlaceStopOrder(askStopLimitOrder, market, user3)

	assert(t, len(ob.stopLimitOrders), 2)

}

func TestMakeStopMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	user3, user2 := createUser()

	price := 40_000.0
	sellOrder := limit.NewLimitOrder(true, 3, user2.ID)
	ob.PlaceLimitOrder(price, sellOrder)

	market := order.MarketUSDT_Fiat

	buyMarketOrder := limit.NewLimitOrder(false, 3, user3.ID)
	matches := ob.PlaceMarketOrder(buyMarketOrder, market)

	assert(t, len(matches), 1)
	assert(t, len(ob.Trades), 1)

	assert(t, ob.Trades[len(ob.Trades)-1].Price, price)

	price2 := 43_000.0
	sellOrder2 := limit.NewLimitOrder(true, 5, 1)
	ob.PlaceLimitOrder(price2, sellOrder2)

	askStopMarketOrderPrice := 43_000.0
	askStopMarketorderStopPrice := 45_000.0
	askStopMarketOrder := order.NewStopOrder(false, false, 5, askStopMarketOrderPrice, askStopMarketorderStopPrice, 3, market)
	ob.PlaceStopOrder(askStopMarketOrder, market, user2)

	assert(t, len(ob.stopMarketOrders), 1)
	assert(t, ob.stopMarketOrders[0].State, order.Pending)

}

func TestCancelStopOrder(t *testing.T) {
	ob := NewOrderbook()

	_, user2 := createUser()

	market := order.MarketUSDT_Fiat
	askStopMarketOrderPrice := 43_000.0
	askStopMarketorderStopPrice := 45_000.0
	askStopMarketOrder := order.NewStopOrder(false, false, 5, askStopMarketOrderPrice, askStopMarketorderStopPrice, user2.ID, market)
	ob.PlaceStopOrder(askStopMarketOrder, market, user2)

	assert(t, len(ob.stopMarketOrders), 1)
	assert(t, ob.stopMarketOrders[len(ob.stopMarketOrders)-1].State, order.Pending)

	ob.CancelStopOrder(askStopMarketOrder)
	assert(t, ob.stopMarketOrders[len(ob.stopMarketOrders)-1].State, order.Canceled)

}
