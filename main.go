package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/client"
	"github.com/mahdifardi/cryptocurrencyExchange/config"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/server"
)

const (
	maxOrders = 3
)

var (
	tick = 2 * time.Second
)

func marketOrderPlacer(config config.Config, c *client.Client, market order.Market) {
	ticker := time.NewTicker(5 * time.Second)

	for {

		newmarketSellOrder := &client.PlaceOrderParams{
			UserId: config.User1ID,
			Bid:    false,
			Size:   1000.0,
			Market: market,
		}

		orderResp, err := c.PlaceMarketOrder(newmarketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		marketSellOrder := &client.PlaceOrderParams{
			UserId: config.User3ID,
			Bid:    false,
			Size:   100.0,
			Market: market,
		}

		orderResp, err = c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserId: config.User3ID,
			Bid:    true,
			Size:   100.0,
			Market: market,
		}

		orderResp, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		<-ticker.C
	}
}

func makeMarketSimple(config config.Config, c *client.Client, market order.Market) {
	ticker := time.NewTicker(tick)

	for {

		trades, err := c.GetTrades(market)
		if err != nil {
			panic(err)
		}

		if len(trades) > 0 {

			fmt.Printf("exchange %s price => %.2f\n", market, trades[len(trades)-1].Price)
		}

		orders, err := c.GetOrders(config.User2ID)
		if err != nil {
			log.Println(err)
		}

		bestBidPrice, err := c.GetBestBid(market)
		if err != nil {
			panic(err)
		}

		bestAskPrice, err := c.GetBestAsk(market)
		if err != nil {
			panic(err)
		}

		spread := math.Abs(bestBidPrice - bestAskPrice)
		fmt.Printf("exchange spread %s : %v\n", market, spread)

		//place the bids
		if len(orders.Orders[market].Bids) < maxOrders {

			bidLimit := &client.PlaceOrderParams{
				UserId: config.User2ID,
				Bid:    true,
				Price:  bestBidPrice + 100,
				Size:   1_000.0,
				Market: market,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp.OrderId)
			}

			// myBids[bidLimit.Price] = bidOrderResp.OrderId
		}

		// place the asks
		if len(orders.Orders[order.MarketETH].Asks) < maxOrders {

			askLimit := &client.PlaceOrderParams{
				UserId: config.User2ID,
				Bid:    false,
				Price:  bestAskPrice - 100,
				Size:   1_000.0,
				Market: market,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp.OrderId)
			}

			// myAsks[askLimit.Price] = askOrderResp.OrderId
		}

		// fmt.Printf("best ask price %s market: %v\n", market, bestAskPrice)
		// fmt.Printf("best bid price %s market: %v\n", market, bestBidPrice)

		<-ticker.C

	}
}

func seedMarket(config config.Config, c *client.Client, market order.Market) error {
	ask := &client.PlaceOrderParams{
		UserId: config.User1ID,
		Bid:    false,
		Price:  10_000.0,
		Size:   10_000.0,
		Market: market,
	}

	bid := &client.PlaceOrderParams{
		UserId: config.User1ID,
		Bid:    true,
		Price:  9_000.0,
		Size:   10_000.0,
		Market: market,
	}

	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}

	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}

	return nil

}

func main() {
	config, err := config.LoadConfig("config/config.json")
	fmt.Println(config)
	if err != nil {
		log.Fatalf("config file error: %v", err)
	}
	go server.StartServer(config)

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	if err := seedMarket(config, c, order.MarketETH); err != nil {
		panic(err)
	}
	if err := seedMarket(config, c, order.MarketBTC); err != nil {
		panic(err)
	}
	if err := seedMarket(config, c, order.MarketUSDT); err != nil {
		panic(err)
	}

	go makeMarketSimple(config, c, order.MarketETH)
	go makeMarketSimple(config, c, order.MarketBTC)
	go makeMarketSimple(config, c, order.MarketUSDT)

	time.Sleep(1 * time.Second)

	go marketOrderPlacer(config, c, order.MarketETH)
	go marketOrderPlacer(config, c, order.MarketBTC)
	go marketOrderPlacer(config, c, order.MarketUSDT)

	select {}
}
