package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
	"github.com/mahdifardi/cryptocurrencyExchange/server"
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
	Price float64
	Size  float64
}

func (c *Client) GetTrades(market string) ([]*orderbook.Trade, error) {
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

func (c *Client) GetOrders(userId int64) (*server.GetOrdersResponse, error) {
	e := fmt.Sprintf("%s/order/%d", Endpoint, userId)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	orders := server.GetOrdersResponse{
		Asks: []server.Order{},
		Bids: []server.Order{},
	}

	if err = json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	return &orders, nil

}

func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserId,
		Type:   server.MarketOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Market: server.MarketETH,
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

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	_ = resp
	// fmt.Printf("%+v", response)

	return placeOrderResponse, nil

}

func (c *Client) GetBestBid() (float64, error) {
	bestPrice := 0.0
	e := fmt.Sprintf("%s/book/ETH/bid", Endpoint)

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return bestPrice, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return bestPrice, err
	}
	priceResp := &server.PriceResponse{}

	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return bestPrice, err
	}

	bestPrice = priceResp.Price

	return bestPrice, nil

}

func (c *Client) GetBestAsk() (float64, error) {
	bestPrice := 0.0
	e := fmt.Sprintf("%s/book/ETH/ask", Endpoint)

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return bestPrice, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return bestPrice, err
	}
	priceResp := &server.PriceResponse{}

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

func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserId,
		Type:   server.LimitOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.MarketETH,
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

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	_ = resp
	// fmt.Printf("%+v", response)

	return placeOrderResponse, nil
}
