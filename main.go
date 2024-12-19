package main

import (
	"time"

	"github.com/mahdifardi/cryptocurrencyExchange/client"
	"github.com/mahdifardi/cryptocurrencyExchange/server"
)

func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	for {
		limitParams1 := &client.PlaceOrderParams{
			UserId: 8888,
			Bid:    false,
			Price:  10_000.0,
			Size:   1_000_000.0,
		}

		_, err := c.PlaceLimitOrder(limitParams1)
		if err != nil {
			panic(err)
		}

		limitParams2 := &client.PlaceOrderParams{
			UserId: 8888,
			Bid:    false,
			Price:  5_000.0,
			Size:   1_000_000.0,
		}

		_, err = c.PlaceLimitOrder(limitParams2)
		if err != nil {
			panic(err)
		}

		//fmt.Println("limit order ffrom the calient => ", resp.OrderId)

		marketParams := &client.PlaceOrderParams{
			UserId: 9999,
			Bid:    true,
			Size:   1_500_000.0,
		}

		_, err = c.PlaceMarketOrder(marketParams)
		if err != nil {
			panic(err)
		}

		//	fmt.Println("market order ffrom the calient => ", resp.OrderId)

		time.Sleep(1 * time.Second)

	}

	select {}
}
