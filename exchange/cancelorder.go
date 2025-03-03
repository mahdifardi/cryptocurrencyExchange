package exchange

import (
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
)

type CancelOrderResponse struct {
	Msg string
}

func (ex *Exchange) CancelOrder(c echo.Context) error {
	// market := c.Param("market")
	var market order.Market
	if err := json.NewDecoder(c.Request().Body).Decode(&market); err != nil {
		return err
	}
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH_Fiat {
		ob = ex.Orderbook[order.MarketETH_Fiat]
	} else if order.Market(market) == order.MarketETH_USDT {
		ob = ex.Orderbook[order.MarketETH_USDT]
	} else if order.Market(market) == order.MarketBTC_Fiat {
		ob = ex.Orderbook[order.MarketBTC_Fiat]
	} else if order.Market(market) == order.MarketBTC_USDT {
		ob = ex.Orderbook[order.MarketBTC_USDT]
	} else if order.Market(market) == order.MarketUSDT_Fiat {
		ob = ex.Orderbook[order.MarketUSDT_Fiat]
	} else {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "Market not supported",
		})
	}

	o, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "order not found",
		})
	}

	user, ok := ex.Users[o.UserId]
	if !ok {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "user of order not found",
		})
	}

	ob.CancelOrder(o)
	log.Println("order canceled id => ", id)

	if o.Bid {
		quoteAmount := new(big.Int).Mul(big.NewInt(int64(o.Limit.Price)), big.NewInt(int64(o.Size)))
		userQuoteBalance := user.AssetBalances[order.Asset(market.Quote)]
		userQuoteBalance.ReservedBalance = new(big.Int).Sub(userQuoteBalance.ReservedBalance, quoteAmount)
		userQuoteBalance.AvailableBalance = new(big.Int).Add(userQuoteBalance.AvailableBalance, quoteAmount)
	} else {
		baseAmount := big.NewInt(int64(o.Size))
		userBaseBalance := user.AssetBalances[order.Asset(market.Base)]
		userBaseBalance.ReservedBalance = new(big.Int).Sub(userBaseBalance.ReservedBalance, baseAmount)
		userBaseBalance.AvailableBalance = new(big.Int).Add(userBaseBalance.AvailableBalance, baseAmount)
	}

	return c.JSON(http.StatusOK, CancelOrderResponse{
		Msg: "order canceled",
	})

}

func (ex *Exchange) CancelStopLimitOrder(c echo.Context) error {
	// market := c.Param("market")
	var market order.Market
	if err := json.NewDecoder(c.Request().Body).Decode(&market); err != nil {
		return err
	}

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH_Fiat {
		ob = ex.Orderbook[order.MarketETH_Fiat]
	} else if order.Market(market) == order.MarketETH_USDT {
		ob = ex.Orderbook[order.MarketETH_USDT]
	} else if order.Market(market) == order.MarketBTC_Fiat {
		ob = ex.Orderbook[order.MarketBTC_Fiat]
	} else if order.Market(market) == order.MarketBTC_USDT {
		ob = ex.Orderbook[order.MarketBTC_USDT]
	} else if order.Market(market) == order.MarketUSDT_Fiat {
		ob = ex.Orderbook[order.MarketUSDT_Fiat]
	} else {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "Market not supported",
		})
	}
	for _, stopLimitOrder := range ob.StopLimits() {
		if stopLimitOrder.ID == int64(id) && stopLimitOrder.State != order.Canceled {
			user, ok := ex.Users[stopLimitOrder.UserID]
			if !ok {
				return c.JSON(http.StatusBadRequest, CancelOrderResponse{
					Msg: "user stoplimitorder not found",
				})
			}
			ob.CancelStopOrder(stopLimitOrder)

			if stopLimitOrder.Bid {
				quoteAmount := new(big.Int).Mul(big.NewInt(int64(stopLimitOrder.Price)), big.NewInt(int64(stopLimitOrder.Size)))
				userQuoteBalance := user.AssetBalances[order.Asset(market.Quote)]
				userQuoteBalance.ReservedBalance = new(big.Int).Sub(userQuoteBalance.ReservedBalance, quoteAmount)
				userQuoteBalance.AvailableBalance = new(big.Int).Add(userQuoteBalance.AvailableBalance, quoteAmount)
			} else {
				baseAmount := big.NewInt(int64(stopLimitOrder.Size))
				userBaseBalance := user.AssetBalances[order.Asset(market.Base)]
				userBaseBalance.ReservedBalance = new(big.Int).Sub(userBaseBalance.ReservedBalance, baseAmount)
				userBaseBalance.AvailableBalance = new(big.Int).Add(userBaseBalance.AvailableBalance, baseAmount)
			}

			return c.JSON(http.StatusOK, CancelOrderResponse{
				Msg: "stop limit order canceled",
			})
		}
	}

	return c.JSON(http.StatusBadRequest, CancelOrderResponse{
		Msg: "stop limit order not found",
	})
}

func (ex *Exchange) CancelStopMarketOrder(c echo.Context) error {
	// market := c.Param("market")
	var market order.Market
	if err := json.NewDecoder(c.Request().Body).Decode(&market); err != nil {
		return err
	}

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH_Fiat {
		ob = ex.Orderbook[order.MarketETH_Fiat]
	} else if order.Market(market) == order.MarketETH_USDT {
		ob = ex.Orderbook[order.MarketETH_USDT]
	} else if order.Market(market) == order.MarketBTC_Fiat {
		ob = ex.Orderbook[order.MarketBTC_Fiat]
	} else if order.Market(market) == order.MarketBTC_USDT {
		ob = ex.Orderbook[order.MarketBTC_USDT]
	} else if order.Market(market) == order.MarketUSDT_Fiat {
		ob = ex.Orderbook[order.MarketUSDT_Fiat]
	} else {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "Market not supported",
		})
	}

	for _, stopMarketOrder := range ob.StopMarkets() {
		if stopMarketOrder.ID == int64(id) && stopMarketOrder.State != order.Canceled {
			user, ok := ex.Users[stopMarketOrder.UserID]
			if !ok {
				return c.JSON(http.StatusBadRequest, CancelOrderResponse{
					Msg: "user of stopMarketOrder not found",
				})
			}
			ob.CancelStopOrder(stopMarketOrder)
			if stopMarketOrder.Bid {
				quoteAmount := new(big.Int).Mul(big.NewInt(int64(stopMarketOrder.Price)), big.NewInt(int64(stopMarketOrder.Size)))
				userQuoteBalance := user.AssetBalances[order.Asset(market.Quote)]
				userQuoteBalance.ReservedBalance = new(big.Int).Sub(userQuoteBalance.ReservedBalance, quoteAmount)
				userQuoteBalance.AvailableBalance = new(big.Int).Add(userQuoteBalance.AvailableBalance, quoteAmount)
			} else {
				baseAmount := big.NewInt(int64(stopMarketOrder.Size))
				userBaseBalance := user.AssetBalances[order.Asset(market.Base)]
				userBaseBalance.ReservedBalance = new(big.Int).Sub(userBaseBalance.ReservedBalance, baseAmount)
				userBaseBalance.AvailableBalance = new(big.Int).Add(userBaseBalance.AvailableBalance, baseAmount)
			}
			return c.JSON(http.StatusOK, CancelOrderResponse{
				Msg: "stop market order canceled",
			})
		}
	}

	return c.JSON(http.StatusBadRequest, CancelOrderResponse{
		Msg: "stop market order not found",
	})
}
