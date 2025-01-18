package exchange

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

type PriceResponse struct {
	Price float64
}

func (ex *Exchange) HandleGetBestBid(c echo.Context) error {
	market := order.Market(c.Param("market"))
	ob := ex.Orderbook[market]

	if len(ob.Bids()) == 0 {
		return fmt.Errorf("the bids are empty")
	}
	bestBidPrice := ob.Bids()[0].Price

	pr := PriceResponse{
		Price: bestBidPrice,
	}
	return c.JSON(http.StatusOK, pr)

}

func (ex *Exchange) HandleGetBestAsk(c echo.Context) error {
	market := order.Market(c.Param("market"))
	ob := ex.Orderbook[market]

	if len(ob.Asks()) == 0 {
		return fmt.Errorf("the asks are empty")
	}
	bestAskPrice := ob.Asks()[0].Price

	pr := PriceResponse{
		Price: bestAskPrice,
	}
	return c.JSON(http.StatusOK, pr)

}
