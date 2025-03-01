package server

import (
	"context"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/config"
	"github.com/mahdifardi/cryptocurrencyExchange/exchange"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

func StartServer(config config.Config) {
	e := echo.New()

	// ethClient, err := createEthClient("http://localh/ost:8545")
	ethClient, err := createEthClient(config.EthServerAddress)

	if err != nil {
		log.Fatal(err)
	}

	btcClient, err := createBtcClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// btcUser1Address := "bcrt1q09umv3yljx5hyn3gptz36q7uc6rmcjxa8wy8ve"
	btcUser1Address := config.BtcUser1Address

	// btcUser2Address := "bcrt1qvqruk47vum9nehcpwhwyzeutcjh2mutwu0efl5"
	btcUser2Address := config.BtcUser2Address

	// btcUser3Address := "bcrt1q4xq3432rt7lj7a3zwazcy0eqgdxrv3gghy4mjh"
	btcUser3Address := config.BtcUser3Address

	ex, err := exchange.NewExchange(config.UstdContractAddress, exchange.ExchangePrivateKey, exchange.ExchangeBTCAdress, ethClient, btcClient)
	if err != nil {
		log.Fatal(err)
	}

	// go ex.TransferBTC(btcClient, btcUser1Address, btcUser2Address, .00002)

	// privKey, _ := crypto.HexToECDSA("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	// usdtAddress := common.HexToAddress("0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab")
	// receiverAddress := common.HexToAddress("0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0")
	// amount := big.NewInt(1000000) // Transfer 1 USDT (assuming 6 decimals)

	// err = ex.TransferUSDT(ethClient, privKey, usdtAddress, receiverAddress, amount)
	// if err != nil {
	// 	log.Fatal("USDT Transfer failed:", err)
	// }

	go ex.ProcessStopLimitOrders(order.MarketETH_Fiat)
	go ex.ProcessStopMarketOrders(order.MarketETH_USDT)

	// pk1 := "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4"
	ethUser1PrivKey := config.EthUser1Address
	user1 := user.NewUser(ethUser1PrivKey, btcUser1Address, config.User1ID)
	ex.Users[user1.ID] = user1

	//market maker
	// pk2 := "b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773"
	ethUser2PrivKey := config.EthUser2Address
	user2 := user.NewUser(ethUser2PrivKey, btcUser2Address, config.User2ID)
	ex.Users[user2.ID] = user2

	// pk3 := "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3"
	ethUser3PrivKey := config.EthUser3Address
	user3 := user.NewUser(ethUser3PrivKey, btcUser3Address, config.User3ID)
	ex.Users[user3.ID] = user3

	// market in body
	e.GET("/trades", ex.HandleGetTrades)

	e.POST("/order", ex.HandlePlaceOrder)
	e.GET("/order/:userId", ex.HandleGetOrders)

	//market in body
	e.DELETE("/order/:id", ex.CancelOrder)
	//market in body
	e.DELETE("/stoplimitorder/:id", ex.CancelStopLimitOrder)
	//market in body
	e.DELETE("/stopmarketorder/:id", ex.CancelStopMarketOrder)

	e.POST("/stoporder", ex.HandlePlaceStopOrder)

	//market in body
	e.GET("/book", ex.HandleGetBook)
	//market in body
	e.GET("/book/bid", ex.HandleGetBestBid)
	//market in body
	e.GET("/book/ask", ex.HandleGetBestAsk)

	address := "0xACa94ef8bD5ffEE41947b4585a84BdA5a3d3DA6E"
	balance, _ := ex.EthClient.BalanceAt(context.Background(), common.HexToAddress(address), nil)

	fmt.Println(balance)

	// e.Start(":3000")
	e.Start(fmt.Sprintf("127.0.0.1:%s", config.ServerPort))

}

func createEthClient(url string) (*ethclient.Client, error) {
	return ethclient.Dial(url)
}

func createBtcClient(config config.Config) (*rpcclient.Client, error) {
	btcConnCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:18332/wallet/regnet_wallet", // fmt.Sprintf("%s:%s%s", config.BtccHostAddress, config.ServerPort, config.BtcWallet),  // Testnet RPC port
		User:         config.BtcUser,                         //"admin",                                // Match rpcuser in bitcoin.conf
		Pass:         config.BtcPass,                         //"admin",                                // Match rpcpassword in bitcoin.conf
		HTTPPostMode: true,                                   // Use HTTP POST mode
		DisableTLS:   true,                                   // Disable TLS for localhost
		Params:       config.BtccParams,                      //"regtest",
		Endpoint:     config.BtcEndpoint,
	}
	return rpcclient.New(btcConnCfg, nil)

}
