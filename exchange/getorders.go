package exchange

import (
	"fmt"
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

	ex.Mu.RLock()
	defer ex.Mu.RUnlock()
	orderBookOrders := ex.Orders[int64(userId)]

	orderResponse := &order.GetOrdersResponse{
		Asks: []order.Order{},
		Bids: []order.Order{},
	}

	// orders := make([]Order, len(orderBookOrders))

	for i := 0; i < len(orderBookOrders); i++ {
		// it could be that ther order is geetting filled even its included in this response we must double check if the limits is not nil

		if orderBookOrders[i].Limit == nil {
			fmt.Println("#################################")
			fmt.Printf("the limmit of the order is nil %+v\n", orderBookOrders[i])
			fmt.Println("#################################")

			continue

		}
		o := order.Order{
			UserID:    orderBookOrders[i].UserId,
			ID:        orderBookOrders[i].ID,
			Price:     orderBookOrders[i].Limit.Price,
			Size:      orderBookOrders[i].Size,
			Bid:       orderBookOrders[i].Bid,
			Timestamp: orderBookOrders[i].Timestamp,
		}
		// orders[i] = o

		if o.Bid {
			orderResponse.Bids = append(orderResponse.Bids, o)
		} else {
			orderResponse.Asks = append(orderResponse.Asks, o)
		}
	}

	// fmt.Printf("%+v", orders)
	return c.JSON(http.StatusOK, orderResponse)
}
