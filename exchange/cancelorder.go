package exchange

import (
	"encoding/json"
	"log"
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

	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, CancelOrderResponse{
			Msg: "order not found",
		})
	}
	ob.CancelOrder(order)

	log.Println("order canceled id => ", id)

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
			ob.CancelStopOrder(stopLimitOrder)
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
			ob.CancelStopOrder(stopMarketOrder)
			return c.JSON(http.StatusOK, CancelOrderResponse{
				Msg: "stop market order canceled",
			})
		}
	}

	return c.JSON(http.StatusBadRequest, CancelOrderResponse{
		Msg: "stop market order not found",
	})
}
