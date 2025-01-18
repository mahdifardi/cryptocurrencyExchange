package exchange

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleGetBook(c echo.Context) error {
	market := c.Param("market")

	ob, ok := ex.Orderbook[order.Market(market)]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "market not found",
		})
	}

	orderBookData := order.OrderBookData{
		TotalBidVolume:   ob.BidTotalVolume(),
		TotalAskVolume:   ob.AskTotalVolume(),
		Asks:             []*order.Order{},
		Bids:             []*order.Order{},
		StopLimitOrders:  []*order.StopOrder{},
		StopMarketOrders: []*order.StopOrder{},
	}

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
			orderBookData.Asks = append(orderBookData.Asks, &o)
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
			orderBookData.Bids = append(orderBookData.Bids, &o)
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
		orderBookData.StopLimitOrders = append(orderBookData.StopLimitOrders, &o)
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
		orderBookData.StopMarketOrders = append(orderBookData.StopMarketOrders, &o)
		// }
	}

	return c.JSON(http.StatusOK, orderBookData)
}
