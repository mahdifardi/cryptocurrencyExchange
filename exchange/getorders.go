package exchange

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleGetOrders(c echo.Context) error {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return err
	}

	// market := c.Param("market")

	ex.Mu.RLock()
	defer ex.Mu.RUnlock()

	// orderBookETHOrders := ex.Orders[order.MarketETH][int64(userId)]
	// orderBookBTCOrders := ex.Orders[order.MarketBTC][int64(userId)]

	// orderResponse := &order.GetOrdersResponse{
	// 	Asks: []order.Order{},
	// 	Bids: []order.Order{},
	// }

	orderResponse := &order.GetOrdersResponse{
		Orders: make(map[order.Market]order.Orders),
	}

	// orderResponse.Orders[order.MarketETH] = order.Orders{}
	// orderResponse.Orders[order.MarketBTC] = order.Orders{}

	// orders := make([]Order, len(orderBookOrders))

	for market, value := range ex.Orders {
		orderResponse.Orders[market] = order.Orders{}
		for _, limitOrders := range value[int64(userId)] {
			if limitOrders.Limit == nil {
				// fmt.Println("#################################")
				// fmt.Printf("the limmit of the order is nil %+v\n", limitOrders)
				// fmt.Println("#################################")

				continue

			}
			o := order.Order{
				UserID:    limitOrders.UserId,
				ID:        limitOrders.ID,
				Price:     limitOrders.Limit.Price,
				Size:      limitOrders.Size,
				Bid:       limitOrders.Bid,
				Timestamp: limitOrders.Timestamp,
			}
			// orders[i] = o

			m := orderResponse.Orders[market]
			if o.Bid {
				m.Bids = append(m.Bids, o)
				orderResponse.Orders[market] = m
			} else {
				// orderResponse.Asks = append(orderResponse.Asks, o)
				m.Asks = append(m.Asks, o)
				orderResponse.Orders[market] = m

			}
		}
	}

	// fmt.Printf("%+v", orders)
	return c.JSON(http.StatusOK, orderResponse)
}
