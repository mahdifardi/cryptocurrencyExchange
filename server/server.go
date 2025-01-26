package server

import (
	"context"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/exchange"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

func StartServer() {
	e := echo.New()

	ethClient, err := createEthClient("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	btcClient, err := createBtcClient()
	if err != nil {
		log.Fatal(err)
	}

	btcUser1Address := "bcrt1q09umv3yljx5hyn3gptz36q7uc6rmcjxa8wy8ve"
	btcUser2Address := "bcrt1qvqruk47vum9nehcpwhwyzeutcjh2mutwu0efl5"

	ex, err := exchange.NewExchange(exchange.ExchangePrivateKey, ethClient, btcClient)
	if err != nil {
		log.Fatal(err)
	}

	go ex.TransferBTC(btcClient, btcUser1Address, btcUser2Address, .00002)

	go ex.ProcessStopLimitOrders(order.MarketETH)
	go ex.ProcessStopMarketOrders(order.MarketETH)

	pk1 := "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4"
	user1 := user.NewUser(pk1, 8888)
	ex.Users[user1.ID] = user1

	pk2 := "b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773"
	user2 := user.NewUser(pk2, 9999)
	ex.Users[user2.ID] = user2

	pk3 := "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3"
	user3 := user.NewUser(pk3, 7777)
	ex.Users[user3.ID] = user3

	e.GET("/trades/:market", ex.HandleGetTrades)

	e.POST("/order", ex.HandlePlaceOrder)
	e.GET("/order/:userId", ex.HandleGetOrders)
	e.DELETE("/order/:id", ex.CancelOrder)
	e.DELETE("/stoplimitorder/:id", ex.CancelStopLimitOrder)
	e.DELETE("/stopmarketorder/:id", ex.CancelStopMarketOrder)

	e.POST("/stoporder", ex.HandlePlaceStopOrder)

	e.GET("/book/:market", ex.HandleGetBook)
	e.GET("/book/:market/bid", ex.HandleGetBestBid)
	e.GET("/book/:market/ask", ex.HandleGetBestAsk)

	address := "0xACa94ef8bD5ffEE41947b4585a84BdA5a3d3DA6E"
	balance, _ := ex.EthClient.BalanceAt(context.Background(), common.HexToAddress(address), nil)

	fmt.Println(balance)

	e.Start(":3000")

}

func createEthClient(url string) (*ethclient.Client, error) {
	return ethclient.Dial(url)
}

func createBtcClient() (*rpcclient.Client, error) {
	btcConnCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:18332/wallet/regnet_wallet", // Testnet RPC port
		User:         "admin",                                // Match rpcuser in bitcoin.conf
		Pass:         "admin",                                // Match rpcpassword in bitcoin.conf
		HTTPPostMode: true,                                   // Use HTTP POST mode
		DisableTLS:   true,                                   // Disable TLS for localhost
		Params:       "regtest",
	}
	return rpcclient.New(btcConnCfg, nil)

}
