package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahdifardi/cryptocurrencyExchange/exchange"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
)

const Endpoint = "http://localhost:3000"

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

type PlaceOrderParams struct {
	UserId int64
	Bid    bool
	// just ffor limit order
	Price  float64
	Size   float64
	Market order.Market
}

func (c *Client) GetTrades(market order.Market) ([]*orderbook.Trade, error) {
	e := fmt.Sprintf("%s/trades/%s", Endpoint, market)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	trades := []*orderbook.Trade{}
	if err = json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, err
	}

	return trades, nil

}

func (c *Client) GetOrders(userId int64) (*order.GetOrdersResponse, error) {
	e := fmt.Sprintf("%s/order/%d", Endpoint, userId)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// orders := order.GetOrdersResponse{
	// 	Asks: []order.Order{},
	// 	Bids: []order.Order{},
	// }

	orders := order.GetOrdersResponse{
		LimitOrders: make(map[order.Market]order.Orders),
		StopOrders:  make(map[order.Market]order.GeneralStopOrders),
	}

	if err = json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	return &orders, nil

}

func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*order.PlaceOrderResponse, error) {
	params := &order.PlaceOrderRequest{
		UserID: p.UserId,
		Type:   order.MarketOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Market: p.Market,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &order.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	_ = resp
	// fmt.Printf("%+v", response)

	return placeOrderResponse, nil

}

func (c *Client) GetBestBid(market order.Market) (float64, error) {
	bestPrice := 0.0
	e := fmt.Sprintf("%s/book/%s/bid", Endpoint, market)

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return bestPrice, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return bestPrice, err
	}
	priceResp := &exchange.PriceResponse{}

	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return bestPrice, err
	}

	bestPrice = priceResp.Price

	return bestPrice, nil

}

func (c *Client) GetBestAsk(market order.Market) (float64, error) {
	bestPrice := 0.0
	e := fmt.Sprintf("%s/book/%s/ask", Endpoint, market)

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return bestPrice, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return bestPrice, err
	}
	priceResp := &exchange.PriceResponse{}

	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return bestPrice, err
	}

	bestPrice = priceResp.Price

	return bestPrice, nil

}

func (c *Client) CancelOrder(orderId int64) error {
	e := fmt.Sprintf("%s/order/%d", Endpoint, orderId)
	req, err := http.NewRequest(http.MethodDelete, e, nil)

	if err != nil {
		return err
	}

	_, err = c.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*order.PlaceOrderResponse, error) {
	params := &order.PlaceOrderRequest{
		UserID: p.UserId,
		Type:   order.LimitOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: p.Market,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &order.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	_ = resp
	// fmt.Printf("%+v", response)

	return placeOrderResponse, nil
}
