package exchange

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) CancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.Orderbook[order.MarketETH]
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
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.Orderbook[order.MarketETH]

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
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.Orderbook[order.MarketETH]

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
