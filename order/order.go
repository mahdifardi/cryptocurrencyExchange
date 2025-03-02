package order

import (
	"time"

	"golang.org/x/exp/rand"
)

const (
	// MarketETH  Market = "ETH"
	// MarketBTC  Market = "BTC"
	// MarketUSDT Market = "USDT"

	MarketOrder OrderType = "Market"
	LimitOrder  OrderType = "Limit"

	Pending   StopOrderState = "PENDING"
	Triggered StopOrderState = "Triggered"
	Canceled  StopOrderState = "Canceled"
)

var Fiat string = "USD"

var AsserETH Asset = "ETH"
var AsserBTC Asset = "BTC"
var AsserUSDT Asset = "USDT"
var AssetFiat Asset = Asset(Fiat)

// market pairs

var MarketETH_Fiat Market = Market{
	Base:  "ETH",
	Quote: Fiat,
}
var MarketETH_USDT Market = Market{
	Base:  "ETH",
	Quote: "USDT",
}

var MarketBTC_Fiat Market = Market{
	Base:  "BTC",
	Quote: Fiat,
}
var MarketBTC_USDT Market = Market{
	Base:  "BTC",
	Quote: "USDT",
}

var MarketUSDT_Fiat Market = Market{
	Base:  "USDT",
	Quote: Fiat,
}

type (
	// Market string
	Asset string

	OrderType string
)

type (
	Market struct {
		Base  string `json:"base"`
		Quote string `json:"quote"`
	}

	MarketString string

	Order struct {
		UserID    int64
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderBookResponse struct {
		Market Market
		State  string
		Data   OrderBookData
	}

	OrderBookData struct {
		TotalBidVolume   float64
		TotalAskVolume   float64
		Asks             []*Order
		Bids             []*Order
		StopLimitOrders  []*StopOrder
		StopMarketOrders []*StopOrder
	}

	StopOrder struct {
		ID        int64
		UserID    int64
		Size      float64
		Bid       bool
		Limit     bool
		Timestamp int64
		StopPrice float64
		Price     float64
		State     StopOrderState
	}

	StopOrders []*StopOrder

	StopOrderState string

	Orders struct {
		Asks []Order `json:"asks"`
		Bids []Order `json:"bids"`
	}

	GeneralStopOrders struct {
		StopLimitOrders  []StopOrder `json:"stopLimitOrders"`
		StopMarketOrders []StopOrder `json:"stopMarketOrders"`
	}

	GetOrdersResponse struct {
		LimitOrders map[MarketString]Orders            `json:"limitOrders"`
		StopOrders  map[MarketString]GeneralStopOrders `json:"stopOrders"`
	}

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	PlaceStopOrderRequest struct {
		UserID    int64
		Bid       bool
		Size      float64
		StopPrice float64
		Price     float64
		Market    Market
		Limit     bool
	}

	MatchedOrder struct {
		UserId int64
		Price  float64
		Size   float64
		ID     int64
	}

	PlaceOrderResponse struct {
		OrderId int64
	}

	PlaceStopOrderResponse struct {
		StopOrderId int64
	}
)

func (m Market) String() string {
	return m.Base + "/" + m.Quote
}

func (so StopOrders) Len() int           { return len(so) }
func (so StopOrders) Swap(i, j int)      { so[i], so[j] = so[j], so[i] }
func (so StopOrders) Less(i, j int) bool { return so[i].StopPrice < so[j].StopPrice }

func NewStopOrder(bid, limit bool, size, price, stopPrice float64, userID int64) *StopOrder {
	return &StopOrder{
		ID:        int64(rand.Intn(1000000)),
		UserID:    userID,
		Size:      size,
		Bid:       bid,
		Limit:     limit,
		Timestamp: time.Now().UnixNano(),
		StopPrice: stopPrice,
		Price:     price,
		State:     Pending,
	}
}
