package exchange

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/config"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
	"github.com/stretchr/testify/assert"
)

func readConfig() config.Config {
	config, err := config.LoadConfig("../config/config.json")
	fmt.Println(config)
	if err != nil {
		log.Fatalf("config file error: %v", err)
	}

	return config
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

func newExchange() *Exchange {
	config := readConfig()

	ethClient, err := createEthClient(config.EthServerAddress)

	if err != nil {
		log.Fatal(err)
	}

	btcClient, err := createBtcClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(config.UstdContractAddress, ExchangePrivateKey, ExchangeBTCAdress, ethClient, btcClient)
	if err != nil {
		log.Fatal(err)
	}

	return ex
}

func TestGetBestBid(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	ob := ex.Orderbook[market]
	bidOrderPrice := 38_000.0
	bidOrderSize := 5
	bidOrderUserId := 4
	Bidorder := limit.NewLimitOrder(true, float64(bidOrderSize), int64(bidOrderUserId))
	ob.PlaceLimitOrder(bidOrderPrice, Bidorder)
	jsonBody, _ := json.Marshal(market)

	req := httptest.NewRequest(http.MethodGet, "/book/bid", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/book/bid")
	// c.SetParamNames("market")
	// c.SetParamValues("ETH")

	err := ex.HandleGetBestBid(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr PriceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, bidOrderPrice, pr.Price)

	// Rainy Path - Empty Orderbook
	ex.Orderbook[market] = orderbook.NewOrderbook() // Reset orderbook to empty
	jsonBody, _ = json.Marshal(market)
	req = httptest.NewRequest(http.MethodGet, "/book/bid", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/book/bid")
	// c.SetParamNames("market")
	// c.SetParamValues("ETH")

	err = ex.HandleGetBestBid(c)
	assert.Error(t, err)
	assert.Equal(t, "the bids are empty", err.Error())
}

func TestGetBestAsk(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	ob := ex.Orderbook[market]
	askOrderPrice := 38_000.0
	askOrderSize := 5
	askOrderUserId := 4
	askOrder := limit.NewLimitOrder(false, float64(askOrderSize), int64(askOrderUserId))
	ob.PlaceLimitOrder(askOrderPrice, askOrder)

	jsonBody, _ := json.Marshal(order.MarketETH_Fiat)
	req := httptest.NewRequest(http.MethodGet, "/book/bid", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/book/bid")
	// c.SetParamNames("market")
	// c.SetParamValues("ETH")

	err := ex.HandleGetBestAsk(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr PriceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, askOrderPrice, pr.Price)

	// Rainy Path - Empty Orderbook
	ex.Orderbook[market] = orderbook.NewOrderbook() // Reset orderbook to empty
	jsonBody, _ = json.Marshal(order.MarketETH_Fiat)
	req = httptest.NewRequest(http.MethodGet, "/book/ask", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/book/ask")
	// c.SetParamNames("market")
	// c.SetParamValues("ETH")

	err = ex.HandleGetBestAsk(c)
	assert.Error(t, err)
	assert.Equal(t, "the asks are empty", err.Error())
}

func TestCancelOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	user := createUser()
	ex.Users[user.ID] = user

	// ob := ex.Orderbook[market]
	askOrderPrice := 38_000.0
	askOrderSize := 5
	// askOrderUserId := 4
	askOrder := limit.NewLimitOrder(false, float64(askOrderSize), user.ID)
	// ob.PlaceLimitOrder(askOrderPrice, askOrder)
	ex.HandlePlaceLimitOrder(market, askOrderPrice, askOrder, user)

	jsonBody, _ := json.Marshal(market)

	tartget := fmt.Sprintf("/order/%v", askOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/order/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(askOrder.ID)))

	err := ex.CancelOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "order canceled", pr.Msg)

	//raniy path order id not exist

	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/order/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/order/:market/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(notExistOrderId))

	err = ex.CancelOrder(c)
	assert.NoError(t, err)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "order not found", prNotExist.Msg)
}

func createUser() *user.User {
	config, err := config.LoadConfig("../config/config.json")
	fmt.Println(config)
	if err != nil {
		log.Fatalf("config file error: %v", err)
	}
	btcUser3Address := config.BtcUser3Address
	ethUser3PrivKey := config.EthUser3Address
	user3 := user.NewUser(ethUser3PrivKey, btcUser3Address, config.User3ID)

	return user3

}
func TestCancelStopLimitOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	user := createUser()
	ex.Users[user.ID] = user

	ob := ex.Orderbook[market]
	stopLimitOrderPrice := 38_000.0
	stopLimitOrderStopPrice := 39_000.0
	stopLimitOrderSize := 5
	// stopLimitOrderUserId := 4
	stopLimitOrder := order.NewStopOrder(false, true, float64(stopLimitOrderSize), stopLimitOrderPrice, stopLimitOrderStopPrice, user.ID)

	ob.PlaceStopOrder(stopLimitOrder, market, user)

	jsonBody, _ := json.Marshal(market)

	tartget := fmt.Sprintf("/stoplimitorder/%v", stopLimitOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(stopLimitOrder.ID)))

	err := ex.CancelStopLimitOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "stop limit order canceled", pr.Msg)

	//raniy path order id not exist
	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/stoplimitorder/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(notExistOrderId)))

	err = ex.CancelStopLimitOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)
	assert.Equal(t, "stop limit order not found", prNotExist.Msg)

	//raniy path Market Not supported

	var id int = 3990
	var notExistMarket = order.Market{
		Base:  "AAA",
		Quote: "BBB",
	}

	notExistMarketJson, _ := json.Marshal(notExistMarket)
	tartget = fmt.Sprintf("/stoplimitorder%v", id)
	req = httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(notExistMarketJson))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(id)))

	err = ex.CancelStopLimitOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prMarketNOtSupport CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prMarketNOtSupport)
	assert.NoError(t, err)
	assert.Equal(t, "Market not supported", prMarketNOtSupport.Msg)
}

func TestCancelStopMarketOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	user := createUser()
	ex.Users[user.ID] = user

	ob := ex.Orderbook[market]
	stopMarketOrderPrice := 38_000.0
	stopMarketOrderStopPrice := 39_000.0
	stopMarketOrderSize := 5
	// stopMarketOrderUserId := 4
	stopMarketOrder := order.NewStopOrder(false, false, float64(stopMarketOrderSize), stopMarketOrderPrice, stopMarketOrderStopPrice, user.ID)

	ob.PlaceStopOrder(stopMarketOrder, market, user)

	jsonBody, _ := json.Marshal(market)

	tartget := fmt.Sprintf("/stopmarketorder/%v", stopMarketOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(stopMarketOrder.ID)))

	err := ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "stop market order canceled", pr.Msg)

	//raniy path order id not exist
	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/stopmarketorder/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(notExistOrderId)))

	err = ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)
	assert.Equal(t, "stop market order not found", prNotExist.Msg)

	//raniy path Market Not supported

	var id int = 3990
	var notExistMarket = order.Market{
		Base:  "AAA",
		Quote: "BBB",
	}

	notExistMarketJson, _ := json.Marshal(notExistMarket)
	tartget = fmt.Sprintf("/stopmarketorder/%v", id)
	req = httptest.NewRequest(http.MethodGet, tartget, bytes.NewReader(notExistMarketJson))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(id)))

	err = ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prMarketNOtSupport CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prMarketNOtSupport)
	assert.NoError(t, err)
	assert.Equal(t, "Market not supported", prMarketNOtSupport.Msg)
}

func TestGetBook(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH_Fiat

	user := createUser()
	ex.Users[user.ID] = user

	ob := ex.Orderbook[market]
	stopMarketOrderPrice := 38_000.0
	stopMarketOrderStopPrice := 39_000.0
	stopMarketOrderSize := 5
	stopMarketOrderUserId := 4
	stopMarketOrder := order.NewStopOrder(false, false, float64(stopMarketOrderSize), stopMarketOrderPrice, stopMarketOrderStopPrice, int64(stopMarketOrderUserId))

	ob.PlaceStopOrder(stopMarketOrder, market, user)

	// tartget := fmt.Sprintf("/book/%v", market)
	jsonBody, _ := json.Marshal(market)
	req := httptest.NewRequest(http.MethodGet, "/book", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/book")
	// c.SetParamNames("market")
	// c.SetParamValues("ETH")

	err := ex.HandleGetBook(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr order.OrderBookResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "supported", pr.State)
	assert.Equal(t, len(pr.Data.StopMarketOrders), 1)

	// rainy path market not supported

	// var notSupportedMarket = "AAA"
	jsonBody, _ = json.Marshal(order.Market{
		Base:  "AAA",
		Quote: "Fiat",
	})
	req = httptest.NewRequest(http.MethodGet, "/book", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/book")
	// c.SetParamNames("market")
	// c.SetParamValues("notSupportedMarket")

	err = ex.HandleGetBook(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prMarketNotSupportd order.OrderBookResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prMarketNotSupportd)
	assert.NoError(t, err)
	assert.Equal(t, "not supported", prMarketNotSupportd.State)
}

func TestGetOrders(t *testing.T) {
	e := echo.New()

	ex := newExchange()

	user1 := createUser()
	ex.Users[user1.ID] = user1

	// limit  order
	jsonBody, _ := json.Marshal(order.PlaceOrderRequest{
		UserID: user1.ID,
		Type:   order.LimitOrder,
		Bid:    true,
		Size:   4,
		Price:  34_000.0,
		Market: order.MarketETH_Fiat,
	})
	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/order")

	err := ex.HandlePlaceOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// stop limit order
	jsonBody, _ = json.Marshal(order.PlaceStopOrderRequest{
		UserID:    user1.ID,
		Bid:       true,
		Size:      3,
		Price:     34_000.0,
		StopPrice: 35_000.0,
		Market:    order.MarketETH_Fiat,
		Limit:     true,
	})
	req = httptest.NewRequest(http.MethodPost, "/stoporder", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoporder")

	err = ex.HandlePlaceStopOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// stop market order
	jsonBody, _ = json.Marshal(order.PlaceStopOrderRequest{
		UserID:    user1.ID,
		Bid:       true,
		Size:      3,
		Price:     34_000.0,
		StopPrice: 35_000.0,
		Market:    order.MarketETH_Fiat,
		Limit:     false,
	})
	req = httptest.NewRequest(http.MethodPost, "/stoporder", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoporder")

	err = ex.HandlePlaceStopOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// get Orders
	tartget := fmt.Sprintf("/order/%v", user1.ID)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/order/:userId")
	c.SetParamNames("userId")
	c.SetParamValues(strconv.Itoa(int(user1.ID)))

	err = ex.HandleGetOrders(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// var pr order.GetOrdersResponse
	pr := order.GetOrdersResponse{
		LimitOrders: make(map[order.MarketString]order.Orders),
		StopOrders:  make(map[order.MarketString]order.GeneralStopOrders),
	}
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pr.LimitOrders[order.MarketString(order.MarketETH_Fiat.String())].Bids))
	assert.Equal(t, 1, len(pr.StopOrders[order.MarketString(order.MarketETH_Fiat.String())].StopLimitOrders))
	assert.Equal(t, 1, len(pr.StopOrders[order.MarketString(order.MarketETH_Fiat.String())].StopMarketOrders))

}

func TestPlaceOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()

	config := readConfig()

	btcUser1Address := config.BtcUser1Address

	btcUser2Address := config.BtcUser2Address

	ethUser1PrivKey := config.EthUser1Address

	ethUser2PrivKey := config.EthUser2Address

	user1 := user.NewUser(ethUser1PrivKey, btcUser1Address, config.User1ID)
	ex.Users[user1.ID] = user1

	user2 := user.NewUser(ethUser2PrivKey, btcUser2Address, config.User2ID)
	ex.Users[user2.ID] = user2

	// limit  order
	jsonBody, _ := json.Marshal(order.PlaceOrderRequest{
		UserID: user1.ID,
		Type:   order.LimitOrder,
		Bid:    true,
		Size:   4,
		Price:  34_000.0,
		Market: order.MarketETH_Fiat,
	})
	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/order")

	err := ex.HandlePlaceOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// assert.Equal(t, ob.AskTotalVolume(), 44)

	//market order
	jsonBody, _ = json.Marshal(order.PlaceOrderRequest{
		UserID: user2.ID,
		Type:   order.MarketOrder,
		Bid:    false,
		Size:   4,
		Price:  34_000.0,
		Market: order.MarketETH_Fiat,
	})
	req = httptest.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/order")

	err = ex.HandlePlaceOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestPlaceStopOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()

	user := createUser()
	ex.Users[user.ID] = user

	// stop limit order
	jsonBody, _ := json.Marshal(order.PlaceStopOrderRequest{
		UserID:    user.ID,
		Bid:       true,
		Size:      3,
		Price:     34_000.0,
		StopPrice: 35_000.0,
		Market:    order.MarketETH_Fiat,
		Limit:     true,
	})
	req := httptest.NewRequest(http.MethodPost, "/stoporder", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/stoporder")

	err := ex.HandlePlaceStopOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// stop market order
	jsonBody, _ = json.Marshal(order.PlaceStopOrderRequest{
		UserID:    user.ID,
		Bid:       true,
		Size:      3,
		Price:     34_000.0,
		StopPrice: 35_000.0,
		Market:    order.MarketETH_Fiat,
		Limit:     false,
	})
	req = httptest.NewRequest(http.MethodPost, "/stoporder", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoporder")

	err = ex.HandlePlaceStopOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetTrades(t *testing.T) {
	e := echo.New()

	ex := newExchange()

	config := readConfig()

	btcUser1Address := config.BtcUser1Address

	btcUser2Address := config.BtcUser2Address

	ethUser1PrivKey := config.EthUser1Address

	ethUser2PrivKey := config.EthUser2Address

	user1 := user.NewUser(ethUser1PrivKey, btcUser1Address, config.User1ID)
	ex.Users[user1.ID] = user1

	user2 := user.NewUser(ethUser2PrivKey, btcUser2Address, config.User2ID)
	ex.Users[user2.ID] = user2

	// limit  order
	jsonBody, _ := json.Marshal(order.PlaceOrderRequest{
		UserID: user1.ID,
		Type:   order.LimitOrder,
		Bid:    true,
		Size:   4,
		Price:  34_000.0,
		Market: order.MarketETH_Fiat,
	})
	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/order")

	err := ex.HandlePlaceOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// assert.Equal(t, ob.AskTotalVolume(), 44)

	//market order
	jsonBody, _ = json.Marshal(order.PlaceOrderRequest{
		UserID: user2.ID,
		Type:   order.MarketOrder,
		Bid:    false,
		Size:   4,
		Price:  34_000.0,
		Market: order.MarketETH_Fiat,
	})
	req = httptest.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/order")

	err = ex.HandlePlaceOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// get trades

	jsonBody, _ = json.Marshal(order.MarketETH_Fiat)
	req = httptest.NewRequest(http.MethodGet, "/trades", bytes.NewBuffer(jsonBody))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/trades")
	// c.SetParamNames("market")
	// c.SetParamValues(string(order.MarketETH_Fiat))

	err = ex.HandleGetTrades(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr []orderbook.Trade
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pr))

}
