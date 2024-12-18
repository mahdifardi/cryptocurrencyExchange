package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
)

const (
	exchangePrivateKey = "9d98cf5774b145b75f4e387b7ee8c763ceb270730052cdcd015ea99f4c1e9652"

	MarketETH Market = "ETH"

	MarketOrder OrderType = "Market"
	LimitOrder  OrderType = "Limit"
)

type (
	Market string

	OrderType string

	Exchange struct {
		Client     *ethclient.Client
		Users      map[int64]*User // orderId => User
		orders     map[int64]int64 //orderid -> userid
		PrivateKey *ecdsa.PrivateKey
		orderbook  map[Market]*orderbook.Orderbook
	}

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		UserID    int64
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderBookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		ID    int64
	}
)

func main() {
	e := echo.New()

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	pk1 := "8a0ff0a7b1624587f2beb88ebaf39e0b99688aa4dec7bdd1500581be80f9314b"
	user1 := NewUser(pk1, 8888)
	ex.Users[user1.ID] = user1

	pk2 := "9d98cf5774b145b75f4e387b7ee8c763ceb270730052cdcd015ea99f4c1e9652"
	user2 := NewUser(pk2, 9999)
	ex.Users[user2.ID] = user2

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.CancelOrder)

	address := "0x9989b28B1FB4628Ec975c99E395E25D802904a3C"
	balance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(address), nil)

	fmt.Println(balance)
	// ctx := context.Background()

	// account := common.HexToAddress("0x8a49Fcf91AbCda383c21025d961d7E9f2A70199b")
	// balance, err := client.BalanceAt(ctx, account, nil)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(balance)

	// privateKey, err := crypto.HexToECDSA("9d98cf5774b145b75f4e387b7ee8c763ceb270730052cdcd015ea99f4c1e9652")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// publicKey := privateKey.Public()
	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	log.Fatal("error casting public key to ECDSA")
	// }

	// fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// gasLimit := uint64(21000) // in units

	// value := big.NewInt(30000000000) // in wei (30 gwei)

	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// toAddress := common.HexToAddress("0x8a49Fcf91AbCda383c21025d961d7E9f2A70199b")
	// tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// chainID, err := client.NetworkID(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = client.SendTransaction(context.Background(), signedTx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// balance, err = client.BalanceAt(ctx, toAddress, nil)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(balance)

	e.Start(":3000")

}

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string, userId int64) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         userId,
		PrivateKey: pk,
	}
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	orderbooks := make(map[Market]*orderbook.Orderbook)

	orderbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		Client:     client,
		Users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		PrivateKey: pk,
		orderbook:  orderbooks,
	}, nil
}

func (ex *Exchange) CancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbook[MarketETH]
	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "order not found",
		})
	}
	ob.CancelOrder(order)

	return c.JSON(http.StatusOK, map[string]any{
		"msg": "order canceled",
	})

}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := c.Param("market")

	ob, ok := ex.orderbook[Market(market)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "market not found",
		})
	}

	orderBookData := OrderBookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserId,
				ID:        order.ID,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserId,
				ID:        order.ID,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderBookData)
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbook[market]

	matches := ob.PlaceMarketOrder(order)
	matchedOreders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	for i := 0; i < len(matchedOreders); i++ {
		id := matches[i].Bid.ID
		if isBid {
			id = matches[i].Ask.ID
		}
		matchedOreders[i] = &MatchedOrder{
			Price: matches[i].Price,
			Size:  matches[i].SizeFilled,
			ID:    id,
		}
	}

	return matches, matchedOreders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbook[market]
	ob.PlaceLimitOrder(price, order)

	// user, ok := ex.Users[order.UserId]
	// if !ok {
	// 	return fmt.Errorf("user not found: %d", user.ID)
	// }
	// // transfffer from user => exchange

	// exchangePublicKey := ex.PrivateKey.Public()
	// exchangePublicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	return fmt.Errorf("error casting public key to ECDSA")
	// }

	// exAddress := crypto.PubkeyToAddress(*exchangePublicKeyECDSA)

	// result := transferETH(ex.Client, user.PrivateKey, exAddress, big.NewInt(int64(order.Size)))

	// return result

	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := placeOrderData.Market
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == LimitOrder {

		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
		return c.JSON(200, map[string]any{
			"msg": "limit order placed",
		})
	}

	if placeOrderData.Type == MarketOrder {
		matches, matchedOreders := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handleMatches(matches); err != nil {
			return err
		}

		return c.JSON(200, map[string]any{
			"matches": matchedOreders,
		})
	}

	return nil
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {

	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.ID)
		}

		toUser, ok := ex.Users[match.Bid.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.ID)
		}

		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		//exchange ffees
		// exchangePublicKey := ex.PrivateKey.Public()
		// exchangePublicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting public key to ECDSA")
		// }

		amount := big.NewInt(int64(match.SizeFilled))

		transferETH(ex.Client, fromUser.PrivateKey, toAddress, amount)

	}

	return nil
}
