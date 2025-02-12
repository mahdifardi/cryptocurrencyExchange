package exchange

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
)

func (ex *Exchange) CancelOrder(c echo.Context) error {
	market := c.Param("market")
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH {
		ob = ex.Orderbook[order.MarketETH]
	} else if order.Market(market) == order.MarketBTC {
		ob = ex.Orderbook[order.MarketBTC]
	} else if order.Market(market) == order.MarketUSDT {
		ob = ex.Orderbook[order.MarketUSDT]
	} else {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "Market not supported",
		})
	}

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

func (ex *Exchange) CancelStopLimitOrder(c echo.Context) error {
	market := c.Param("market")

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH {
		ob = ex.Orderbook[order.MarketETH]
	} else if order.Market(market) == order.MarketBTC {
		ob = ex.Orderbook[order.MarketBTC]
	} else if order.Market(market) == order.MarketUSDT {
		ob = ex.Orderbook[order.MarketUSDT]
	} else {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "Market not supported",
		})
	}
	for _, stopLimitOrder := range ob.StopLimits() {
		if stopLimitOrder.ID == int64(id) && stopLimitOrder.State != order.Canceled {
			ob.CancelStopOrder(stopLimitOrder)
			return c.JSON(http.StatusOK, map[string]any{
				"msg": "stop limit order canceled",
			})
		}
	}

	return c.JSON(http.StatusBadRequest, map[string]any{
		"msg": "stop limit order not found",
	})
}

func (ex *Exchange) CancelStopMarketOrder(c echo.Context) error {
	market := c.Param("market")

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var ob *orderbook.Orderbook
	if order.Market(market) == order.MarketETH {
		ob = ex.Orderbook[order.MarketETH]
	} else if order.Market(market) == order.MarketBTC {
		ob = ex.Orderbook[order.MarketBTC]
	} else if order.Market(market) == order.MarketUSDT {
		ob = ex.Orderbook[order.MarketUSDT]
	} else {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "Market not supported",
		})
	}

	for _, stopMarketOrder := range ob.StopMarkets() {
		if stopMarketOrder.ID == int64(id) && stopMarketOrder.State != order.Canceled {
			ob.CancelStopOrder(stopMarketOrder)
			return c.JSON(http.StatusOK, map[string]any{
				"msg": "stop market order canceled",
			})
		}
	}

	return c.JSON(http.StatusBadRequest, map[string]any{
		"msg": "stop market order not found",
	})
}
