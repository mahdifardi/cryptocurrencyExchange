package exchange

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

func (ex *Exchange) HandlePlaceMarketOrder(market order.Market, newOrder *limit.LimitOrder) ([]limit.Match, []*order.MatchedOrder) {
	ob := ex.Orderbook[market]

	matches := ob.PlaceMarketOrder(newOrder, market)
	matchedOreders := make([]*order.MatchedOrder, len(matches))

	isBid := false
	if newOrder.Bid {
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
		matchedOreders[i] = &order.MatchedOrder{
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
	log.Printf("filled Market order => market [%s] | orderId [%d] | size [%.2f] | average price [%.2f]", market, newOrder.ID, totalSizeFilled, averagePrice)

	return matches, matchedOreders
}

func (ex *Exchange) HandlePlaceLimitOrder(market order.Market, price float64, newOrder *limit.LimitOrder, user *user.User) error {
	ob := ex.Orderbook[market]
	if newOrder.Bid {
		quoteAmount := new(big.Int).Mul(big.NewInt(int64(price)), big.NewInt(int64(newOrder.Size)))
		err := user.ReserveBalance(market, quoteAmount, newOrder.Bid)
		if err != nil {
			return err
		}
		// userQuoteAvailableBalance := user.AssetBalances[order.Asset(market.Quote)].AvailableBalance
		// if userQuoteAvailableBalance.Cmp(quoteAmount) < 0 {
		// 	return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Quote, userQuoteAvailableBalance, quoteAmount)
		// } else {
		// 	userQuoteBalance := user.AssetBalances[order.Asset(market.Quote)]
		// 	userQuoteBalance.AvailableBalance = new(big.Int).Sub(userQuoteBalance.AvailableBalance, quoteAmount)
		// 	userQuoteBalance.ReservedBalance = new(big.Int).Add(userQuoteBalance.ReservedBalance, quoteAmount)
		// 	user.AssetBalances[order.Asset(market.Quote)] = userQuoteBalance
		// }
	} else {
		BaseAmount := big.NewInt(int64(newOrder.Size))
		err := user.ReserveBalance(market, BaseAmount, newOrder.Bid)
		if err != nil {
			return err
		}
		// userBaseBalance := user.AssetBalances[order.Asset(market.Base)].AvailableBalance
		// if userBaseBalance.Cmp(BaseAmount) < 0 {
		// 	return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Quote, userBaseBalance, BaseAmount)
		// } else {
		// 	userBaseBalance := user.AssetBalances[order.Asset(market.Base)]
		// 	userBaseBalance.AvailableBalance = new(big.Int).Sub(userBaseBalance.AvailableBalance, BaseAmount)
		// 	userBaseBalance.ReservedBalance = new(big.Int).Add(userBaseBalance.ReservedBalance, BaseAmount)
		// 	user.AssetBalances[order.Asset(market.Base)] = userBaseBalance
		// }
	}
	ob.PlaceLimitOrder(price, newOrder)

	ex.Mu.Lock()
	// if market == order.MarketETH {

	// 	ex.Orders[order.MarketETH][newOrder.UserId] = append(ex.Orders[order.MarketETH][newOrder.UserId], newOrder)
	// } else if market == order.MarketBTC {
	// 	ex.Orders[order.MarketBTC][newOrder.UserId] = append(ex.Orders[order.MarketBTC][newOrder.UserId], newOrder)

	// }

	ex.LimitOrders[market][newOrder.UserId] = append(ex.LimitOrders[market][newOrder.UserId], newOrder)
	ex.Mu.Unlock()
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

	log.Printf("new limit order => market [%s]type [%t] price [%.2f] size [%.2f]", market, newOrder.Bid, newOrder.Limit.Price, newOrder.Size)

	return nil
}

func (ex *Exchange) HandlePlaceStopOrder(c echo.Context) error {
	var placeStopOrderData order.PlaceStopOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeStopOrderData); err != nil {
		return err
	}

	market := placeStopOrderData.Market
	newOrder := order.NewStopOrder(placeStopOrderData.Bid, placeStopOrderData.Limit, placeStopOrderData.Size, placeStopOrderData.Price, placeStopOrderData.StopPrice, placeStopOrderData.UserID, placeStopOrderData.Market)

	ob := ex.Orderbook[market]

	user, ok := ex.Users[placeStopOrderData.UserID]
	if !ok {
		return fmt.Errorf("user not found error")
	}

	err := ob.PlaceStopOrder(newOrder, market, user)
	if err != nil {
		return err
	}

	ex.StopOrders[market][placeStopOrderData.UserID] = append(ex.StopOrders[market][placeStopOrderData.UserID], newOrder)

	resp := &order.PlaceStopOrderResponse{
		StopOrderId: newOrder.ID,
	}

	return c.JSON(200, resp)
}

func (ex *Exchange) HandlePlaceOrder(c echo.Context) error {
	var placeOrderData order.PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	//Limimmt
	market := placeOrderData.Market
	newOrder := limit.NewLimitOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == order.LimitOrder {
		user, ok := ex.Users[placeOrderData.UserID]
		if !ok {
			return fmt.Errorf("user not found error")
		}

		if err := ex.HandlePlaceLimitOrder(market, placeOrderData.Price, newOrder, user); err != nil {
			return err
		}

	}

	//Market
	if placeOrderData.Type == order.MarketOrder {
		matches, matchedOrders := ex.HandlePlaceMarketOrder(market, newOrder)

		if err := ex.HandleMatches(placeOrderData.Bid, market, matches); err != nil {
			return c.JSON(500, err)
		}

		//delete the orders off the user wwhen filled
		// for _, matchedOreder := range matchedOreders {

		for j := 0; j < len(matchedOrders); j++ {
			ordersForUser := ex.LimitOrders[placeOrderData.Market][matchedOrders[j].UserId]
			// Iterate backwards
			for i := len(ordersForUser) - 1; i >= 0; i-- {
				if ordersForUser[i].IsFilled() && matchedOrders[j].ID == ordersForUser[i].ID {
					// Remove the element at index i
					ordersForUser = append(ordersForUser[:i], ordersForUser[i+1:]...)
				}
			}
			ex.LimitOrders[placeOrderData.Market][matchedOrders[j].UserId] = ordersForUser
		}

		// m := len(matchedOreders)
		// for j := 0; j < m; j++ {
		// 	// userOrders :=  ex.Orders[matchedOreders[j].UserId]
		// 	n := len(ex.Orders[placeOrderData.Market][matchedOreders[j].UserId])
		// 	for i := 0; i < n; i++ {

		// 		// if the size is 0 ew can delete order
		// 		if ex.Orders[placeOrderData.Market][matchedOreders[j].UserId][i].IsFilled() {

		// 			if matchedOreders[j].ID == ex.Orders[placeOrderData.Market][matchedOreders[j].UserId][i].ID {
		// 				// log.Printf("len j:%d, j:%d, len i:%d i:%d", len(matchedOreders), j, len(ex.Orders[placeOrderData.Market][matchedOreders[j].UserId]), i)
		// 				ex.Orders[placeOrderData.Market][matchedOreders[j].UserId][i] = ex.Orders[placeOrderData.Market][matchedOreders[j].UserId][len(ex.Orders[placeOrderData.Market][matchedOreders[j].UserId])-1]
		// 				ex.Orders[placeOrderData.Market][matchedOreders[j].UserId] = ex.Orders[placeOrderData.Market][matchedOreders[j].UserId][:len(ex.Orders[placeOrderData.Market][matchedOreders[j].UserId])-1]
		// 			}
		// 		}
		// 	}

		// }
	}

	resp := &order.PlaceOrderResponse{
		OrderId: newOrder.ID,
	}

	return c.JSON(200, resp)

}
