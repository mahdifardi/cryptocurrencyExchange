package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/client"
	"github.com/mahdifardi/cryptocurrencyExchange/config"
	"github.com/mahdifardi/cryptocurrencyExchange/server"
)

const (
	maxOrders = 3
)

var (
	tick = 2 * time.Second
)

func marketOrderPlacer(config config.Config, c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)

	for {

		newmarketSellOrder := &client.PlaceOrderParams{
			UserId: config.User1ID,
			Bid:    false,
			Size:   1000.0,
		}

		orderResp, err := c.PlaceMarketOrder(newmarketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		marketSellOrder := &client.PlaceOrderParams{
			UserId: config.User3ID,
			Bid:    false,
			Size:   100.0,
		}

		orderResp, err = c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserId: config.User3ID,
			Bid:    true,
			Size:   100.0,
		}

		orderResp, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		<-ticker.C
	}
}

func makeMarketSimple(config config.Config, c *client.Client) {
	ticker := time.NewTicker(tick)

	for {

		trades, err := c.GetTrades("ETH")
		if err != nil {
			panic(err)
		}

		if len(trades) > 0 {

			fmt.Printf("exchange price => %.2f\n", trades[len(trades)-1].Price)
		}

		orders, err := c.GetOrders(config.User2ID)
		if err != nil {
			log.Println(err)
		}

		bestBidPrice, err := c.GetBestBid()
		if err != nil {
			panic(err)
		}

		bestAskPrice, err := c.GetBestAsk()
		if err != nil {
			panic(err)
		}

		spread := math.Abs(bestBidPrice - bestAskPrice)
		fmt.Println("exchange spread", spread)

		//place the bids
		if len(orders.Bids) < maxOrders {

			bidLimit := &client.PlaceOrderParams{
				UserId: config.User2ID,
				Bid:    true,
				Price:  bestBidPrice + 100,
				Size:   1_000.0,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp.OrderId)
			}

			// myBids[bidLimit.Price] = bidOrderResp.OrderId
		}

		// place the asks
		if len(orders.Asks) < maxOrders {

			askLimit := &client.PlaceOrderParams{
				UserId: config.User2ID,
				Bid:    false,
				Price:  bestAskPrice - 100,
				Size:   1_000.0,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp.OrderId)
			}

			// myAsks[askLimit.Price] = askOrderResp.OrderId
		}

		fmt.Println("best ask price", bestAskPrice)
		fmt.Println("best bid price", bestBidPrice)

		<-ticker.C

	}
}

func seedMarket(config config.Config, c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserId: config.User1ID,
		Bid:    false,
		Price:  10_000.0,
		Size:   10_000.0,
	}

	bid := &client.PlaceOrderParams{
		UserId: config.User1ID,
		Bid:    true,
		Price:  9_000.0,
		Size:   10_000.0,
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

	if err := seedMarket(config, c); err != nil {
		panic(err)
	}

	go makeMarketSimple(config, c)
	time.Sleep(1 * time.Second)
	marketOrderPlacer(config, c)

	// for {
	// 	limitParams1 := &client.PlaceOrderParams{
	// 		UserId: 8888,
	// 		Bid:    false,
	// 		Price:  10_000.0,
	// 		Size:   5_000_000.0,
	// 	}

	// 	_, err := c.PlaceLimitOrder(limitParams1)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	limitParams2 := &client.PlaceOrderParams{
	// 		UserId: 8888,
	// 		Bid:    false,
	// 		Price:  9_000.0,
	// 		Size:   500_000.0,
	// 	}

	// 	_, err = c.PlaceLimitOrder(limitParams2)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	buyLimitOrder := &client.PlaceOrderParams{
	// 		UserId: 8888,
	// 		Bid:    true,
	// 		Price:  11_000.0,
	// 		Size:   500_000.0,
	// 	}

	// 	_, err = c.PlaceLimitOrder(buyLimitOrder)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	//fmt.Println("limit order ffrom the calient => ", resp.OrderId)

	// 	//	fmt.Println("market order ffrom the calient => ", resp.OrderId)

	// 	marketParams := &client.PlaceOrderParams{
	// 		UserId: 9999,
	// 		Bid:    true,
	// 		Size:   1_000_000.0,
	// 	}

	// 	_, err = c.PlaceMarketOrder(marketParams)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	bestBidPrice, err := c.GetBestBid()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("best bid price: %.2f\n", bestBidPrice)

	// 	bestAskPrice, err := c.GetBestAsk()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("best ask price: %.2f\n", bestAskPrice)

	// 	time.Sleep(1 * time.Second)

	// }

	select {}
}
