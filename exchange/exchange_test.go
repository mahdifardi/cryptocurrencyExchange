package exchange

import (
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
	market := order.MarketETH

	ob := ex.Orderbook[market]
	bidOrderPrice := 38_000.0
	bidOrderSize := 5
	bidOrderUserId := 4
	Bidorder := limit.NewLimitOrder(true, float64(bidOrderSize), int64(bidOrderUserId))
	ob.PlaceLimitOrder(bidOrderPrice, Bidorder)

	req := httptest.NewRequest(http.MethodGet, "/book/ETH/bid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/book/:market/bid")
	c.SetParamNames("market")
	c.SetParamValues("ETH")

	err := ex.HandleGetBestBid(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr PriceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, bidOrderPrice, pr.Price)

	// Rainy Path - Empty Orderbook
	ex.Orderbook[market] = orderbook.NewOrderbook() // Reset orderbook to empty
	req = httptest.NewRequest(http.MethodGet, "/book/ETH/bid", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/book/:market/bid")
	c.SetParamNames("market")
	c.SetParamValues("ETH")

	err = ex.HandleGetBestBid(c)
	assert.Error(t, err)
	assert.Equal(t, "the bids are empty", err.Error())
}

func TestGetBestAsk(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH

	ob := ex.Orderbook[market]
	askOrderPrice := 38_000.0
	askOrderSize := 5
	askOrderUserId := 4
	askOrder := limit.NewLimitOrder(false, float64(askOrderSize), int64(askOrderUserId))
	ob.PlaceLimitOrder(askOrderPrice, askOrder)

	req := httptest.NewRequest(http.MethodGet, "/book/ETH/bid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/book/:market/ask")
	c.SetParamNames("market")
	c.SetParamValues("ETH")

	err := ex.HandleGetBestAsk(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr PriceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, askOrderPrice, pr.Price)

	// Rainy Path - Empty Orderbook
	ex.Orderbook[market] = orderbook.NewOrderbook() // Reset orderbook to empty
	req = httptest.NewRequest(http.MethodGet, "/book/ETH/ask", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/book/:market/ask")
	c.SetParamNames("market")
	c.SetParamValues("ETH")

	err = ex.HandleGetBestAsk(c)
	assert.Error(t, err)
	assert.Equal(t, "the asks are empty", err.Error())
}

func TestCancelOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH

	ob := ex.Orderbook[market]
	askOrderPrice := 38_000.0
	askOrderSize := 5
	askOrderUserId := 4
	askOrder := limit.NewLimitOrder(false, float64(askOrderSize), int64(askOrderUserId))
	ob.PlaceLimitOrder(askOrderPrice, askOrder)

	tartget := fmt.Sprintf("/order/ETH/%v", askOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/order/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(int(askOrder.ID)))

	err := ex.CancelOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "order canceled", pr.Msg)

	//raniy path order id not exist

	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/order/ETH/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/order/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(notExistOrderId))

	err = ex.CancelOrder(c)
	assert.NoError(t, err)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "order not found", prNotExist.Msg)
}

func TestCancelStopLimitOrder(t *testing.T) {
	e := echo.New()

	ex := newExchange()
	market := order.MarketETH

	ob := ex.Orderbook[market]
	stopLimitOrderPrice := 38_000.0
	stopLimitOrderStopPrice := 39_000.0
	stopLimitOrderSize := 5
	stopLimitOrderUserId := 4
	stopLimitOrder := order.NewStopOrder(false, true, float64(stopLimitOrderSize), stopLimitOrderPrice, stopLimitOrderStopPrice, int64(stopLimitOrderUserId))

	ob.PlaceStopOrder(stopLimitOrder)

	tartget := fmt.Sprintf("/stoplimitorder/ETH/%v", stopLimitOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(int(stopLimitOrder.ID)))

	err := ex.CancelStopLimitOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "stop limit order canceled", pr.Msg)

	//raniy path order id not exist
	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/stoplimitorder/ETH/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(int(notExistOrderId)))

	err = ex.CancelStopLimitOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)
	assert.Equal(t, "stop limit order not found", prNotExist.Msg)

	//raniy path Market Not supported

	var id int = 3990
	var notExistMarket string = "AAA"
	tartget = fmt.Sprintf("/stoplimitorder/ETH/%v", id)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stoplimitorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues(notExistMarket, strconv.Itoa(int(id)))

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
	market := order.MarketETH

	ob := ex.Orderbook[market]
	stopMarketOrderPrice := 38_000.0
	stopMarketOrderStopPrice := 39_000.0
	stopMarketOrderSize := 5
	stopMarketOrderUserId := 4
	stopMarketOrder := order.NewStopOrder(false, false, float64(stopMarketOrderSize), stopMarketOrderPrice, stopMarketOrderStopPrice, int64(stopMarketOrderUserId))

	ob.PlaceStopOrder(stopMarketOrder)

	tartget := fmt.Sprintf("/stopmarketorder/ETH/%v", stopMarketOrder.ID)
	req := httptest.NewRequest(http.MethodGet, tartget, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(int(stopMarketOrder.ID)))

	err := ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pr CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &pr)
	assert.NoError(t, err)
	assert.Equal(t, "stop market order canceled", pr.Msg)

	//raniy path order id not exist
	var notExistOrderId int = 101010

	tartget = fmt.Sprintf("/stopmarketorder/ETH/%v", notExistOrderId)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues("ETH", strconv.Itoa(int(notExistOrderId)))

	err = ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prNotExist CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prNotExist)
	assert.NoError(t, err)
	assert.Equal(t, "stop market order not found", prNotExist.Msg)

	//raniy path Market Not supported

	var id int = 3990
	var notExistMarket string = "AAA"
	tartget = fmt.Sprintf("/stopmarketorder/ETH/%v", id)
	req = httptest.NewRequest(http.MethodGet, tartget, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/stopmarketorder/:market/:id")
	c.SetParamNames("market", "id")
	c.SetParamValues(notExistMarket, strconv.Itoa(int(id)))

	err = ex.CancelStopMarketOrder(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prMarketNOtSupport CancelOrderResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prMarketNOtSupport)
	assert.NoError(t, err)
	assert.Equal(t, "Market not supported", prMarketNOtSupport.Msg)
}
