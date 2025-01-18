package exchange

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleGetTrades(c echo.Context) error {
	market := order.Market(c.Param("market"))
	ob, ok := ex.Orderbook[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, APIError{Error: "orderbook not found"})
	}

	return c.JSON(http.StatusOK, ob.Trades)

}
