package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
		mu         sync.RWMutex
		Client     *ethclient.Client
		Users      map[int64]*User              // orderId => User
		Orders     map[int64][]*orderbook.Order // ordeers map a user to his orders
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
		UserId int64
		Price  float64
		Size   float64
		ID     int64
	}
)

func StartServer() {
	e := echo.New()

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	pk1 := "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4"
	user1 := NewUser(pk1, 8888)
	ex.Users[user1.ID] = user1

	pk2 := "b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773"
	user2 := NewUser(pk2, 9999)
	ex.Users[user2.ID] = user2

	pk3 := "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3"
	user3 := NewUser(pk3, 7777)
	ex.Users[user3.ID] = user3

	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/order/:userId", ex.handleGetOrders)
	e.DELETE("/order/:id", ex.CancelOrder)

	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bid", ex.handleGetBestBid)
	e.GET("/book/:market/ask", ex.handleGetBestAsk)

	address := "0xACa94ef8bD5ffEE41947b4585a84BdA5a3d3DA6E"
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
		Orders:     make(map[int64][]*orderbook.Order),
		PrivateKey: pk,
		orderbook:  orderbooks,
	}, nil
}

type PriceResponse struct {
	Price float64
}

func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return err
	}

	ex.mu.Lock()
	defer ex.mu.Unlock()
	orderBookOrders := ex.Orders[int64(userId)]

	orders := make([]Order, len(orderBookOrders))

	for i := 0; i < len(orderBookOrders); i++ {
		o := Order{
			UserID:    orderBookOrders[i].UserId,
			ID:        orderBookOrders[i].ID,
			Price:     orderBookOrders[i].Limit.Price,
			Size:      orderBookOrders[i].Size,
			Bid:       orderBookOrders[i].Bid,
			Timestamp: orderBookOrders[i].Timestamp,
		}
		orders[i] = o
	}

	// fmt.Printf("%+v", orders)
	return c.JSON(http.StatusOK, orders)
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbook[market]

	if len(ob.Bids()) == 0 {
		return fmt.Errorf("the bids are empty")
	}
	bestBidPrice := ob.Bids()[0].Price

	pr := PriceResponse{
		Price: bestBidPrice,
	}
	return c.JSON(http.StatusOK, pr)

}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbook[market]

	if len(ob.Asks()) == 0 {
		return fmt.Errorf("the asks are empty")
	}
	bestAskPrice := ob.Asks()[0].Price

	pr := PriceResponse{
		Price: bestAskPrice,
	}
	return c.JSON(http.StatusOK, pr)

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

	log.Println("order canceled id => ", id)

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
	sumPrice := 0.0
	totalSizeFilled := 0.0
	for i := 0; i < len(matchedOreders); i++ {
		id := matches[i].Bid.ID
		userId := matches[i].Bid.UserId
		if isBid {
			id = matches[i].Ask.ID
			userId = matches[i].Ask.UserId

		}
		matchedOreders[i] = &MatchedOrder{
			// UserId: order.UserId,
			UserId: userId,
			Price:  matches[i].Price,
			Size:   matches[i].SizeFilled,
			ID:     id,
		}
		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price
	}
	averagePrice := sumPrice / float64((len(matchedOreders)))
	log.Printf("filled Market order =>%d | size [%.2f] | average price [%.2f]", order.ID, totalSizeFilled, averagePrice)

	return matches, matchedOreders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbook[market]
	ob.PlaceLimitOrder(price, order)

	ex.mu.Lock()
	ex.Orders[order.UserId] = append(ex.Orders[order.UserId], order)
	ex.mu.Unlock()
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

	log.Printf("new limit order =>type [%t] price [%.2f] size [%.2f]", order.Bid, order.Limit.Price, order.Size)

	return nil
}

type PlaceOrderResponse struct {
	OrderId int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	//Limimmt
	market := placeOrderData.Market
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == LimitOrder {

		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}

	}

	//Market
	if placeOrderData.Type == MarketOrder {
		matches, matchedOreders := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handleMatches(matches); err != nil {
			return err
		}

		//delete the orders off the user wwhen filled
		// for _, matchedOreder := range matchedOreders {
		for j := 0; j < len(matchedOreders); j++ {
			// userOrders :=  ex.Orders[matchedOreders[j].UserId]
			for i := 0; i < len(ex.Orders[matchedOreders[j].UserId]); i++ {

				// if the size is 0 ew can delete order
				if ex.Orders[matchedOreders[j].UserId][i].IsFilled() {

					if matchedOreders[j].ID == ex.Orders[matchedOreders[j].UserId][i].ID {
						ex.Orders[matchedOreders[j].UserId][i] = ex.Orders[matchedOreders[j].UserId][len(ex.Orders[matchedOreders[j].UserId])-1]
						ex.Orders[matchedOreders[j].UserId] = ex.Orders[matchedOreders[j].UserId][:len(ex.Orders[matchedOreders[j].UserId])-1]
					}
				}
			}

		}
	}

	resp := &PlaceOrderResponse{
		OrderId: order.ID,
	}
	return c.JSON(200, resp)

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

func transferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()

	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	return nil
}
