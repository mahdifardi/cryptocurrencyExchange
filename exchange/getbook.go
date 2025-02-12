package exchange

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleGetBook(c echo.Context) error {
	market := c.Param("market")

	var response order.OrderBookResponse
	response.Market = order.Market(market)

	ob, ok := ex.Orderbook[order.Market(market)]

	if !ok {
		response.State = "not supported"
		return c.JSON(http.StatusBadRequest, response)
	}

	response.State = "supported"

	response.Data.TotalBidVolume = ob.BidTotalVolume()
	response.Data.TotalAskVolume = ob.AskTotalVolume()
	response.Data.Asks = []*order.Order{}
	response.Data.Bids = []*order.Order{}
	response.Data.StopLimitOrders = []*order.StopOrder{}
	response.Data.StopMarketOrders = []*order.StopOrder{}

	for _, limit := range ob.Asks() {
		for _, lOrder := range limit.Orders {
			o := order.Order{
				UserID:    lOrder.UserId,
				ID:        lOrder.ID,
				Price:     lOrder.Limit.Price,
				Size:      lOrder.Size,
				Bid:       lOrder.Bid,
				Timestamp: lOrder.Timestamp,
			}
			response.Data.Asks = append(response.Data.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, lOrder := range limit.Orders {
			o := order.Order{
				UserID:    lOrder.UserId,
				ID:        lOrder.ID,
				Price:     lOrder.Limit.Price,
				Size:      lOrder.Size,
				Bid:       lOrder.Bid,
				Timestamp: lOrder.Timestamp,
			}
			response.Data.Bids = append(response.Data.Bids, &o)
		}
	}

	for _, stopLimitOrder := range ob.StopLimits() {
		// if stopLimitOrder.State == orderbook.Pending {
		o := order.StopOrder{
			ID:        stopLimitOrder.ID,
			UserID:    stopLimitOrder.UserID,
			Size:      stopLimitOrder.Size,
			Bid:       stopLimitOrder.Bid,
			Limit:     stopLimitOrder.Limit,
			Timestamp: stopLimitOrder.Timestamp,
			StopPrice: stopLimitOrder.StopPrice,
			Price:     stopLimitOrder.Price,
			State:     stopLimitOrder.State,
		}
		response.Data.StopLimitOrders = append(response.Data.StopLimitOrders, &o)
		// }
	}

	for _, stopMarketOrder := range ob.StopMarkets() {
		// if stopMarketOrder.State == orderbook.Pending {
		o := order.StopOrder{
			ID:        stopMarketOrder.ID,
			UserID:    stopMarketOrder.UserID,
			Size:      stopMarketOrder.Size,
			Bid:       stopMarketOrder.Bid,
			Limit:     stopMarketOrder.Bid,
			Timestamp: stopMarketOrder.Timestamp,
			StopPrice: stopMarketOrder.StopPrice,
			Price:     stopMarketOrder.Price,
			State:     stopMarketOrder.State,
		}
		response.Data.StopMarketOrders = append(response.Data.StopMarketOrders, &o)
		// }
	}

	return c.JSON(http.StatusOK, response)
}
