package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/client"
	"github.com/mahdifardi/cryptocurrencyExchange/server"
)

const (
	maxOrders = 3
)

var (
	tick = 2 * time.Second

	myAsks = make(map[float64]int64) // price => orderId
	myBids = make(map[float64]int64) // price => orderId
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		marketSellOrder := &client.PlaceOrderParams{
			UserId: 7777,
			Bid:    false,
			Size:   1_000.0,
		}

		orderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserId: 7777,
			Bid:    true,
			Size:   1_000.0,
		}

		orderResp, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp.OrderId)
		}

		<-ticker.C
	}
}

func makeMarketSimple(c *client.Client) {
	ticker := time.NewTicker(tick)

	for {

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
		if len(myBids) < maxOrders {

			bidLimit := &client.PlaceOrderParams{
				UserId: 9999,
				Bid:    true,
				Price:  bestBidPrice + 100,
				Size:   1_000.0,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp.OrderId)
			}

			myBids[bidLimit.Price] = bidOrderResp.OrderId
		}

		// place the asks
		if len(myAsks) < maxOrders {

			askLimit := &client.PlaceOrderParams{
				UserId: 9999,
				Bid:    false,
				Price:  bestAskPrice - 100,
				Size:   1_000.0,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp.OrderId)
			}

			myAsks[askLimit.Price] = askOrderResp.OrderId
		}

		fmt.Println("best ask price", bestAskPrice)
		fmt.Println("best bid price", bestBidPrice)

		<-ticker.C

	}
}

func seeMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserId: 8888,
		Bid:    false,
		Price:  10_000.0,
		Size:   1_000_000.0,
	}

	bid := &client.PlaceOrderParams{
		UserId: 8888,
		Bid:    true,
		Price:  9_000.0,
		Size:   1_000_000.0,
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
	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	if err := seeMarket(c); err != nil {
		panic(err)
	}

	go makeMarketSimple(c)
	time.Sleep(1 * time.Second)
	marketOrderPlacer(c)
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
