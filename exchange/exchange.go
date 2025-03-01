package exchange

import (
	"crypto/ecdsa"
	"sync"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/orderbook"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

const (
	ExchangePrivateKey = "9d98cf5774b145b75f4e387b7ee8c763ceb270730052cdcd015ea99f4c1e9652"
	ExchangeBTCAdress  = "bcrt1qvqruk47vum9nehcpwhwyzeutcjh2mutwu0efl5"
)

type (
	APIError struct {
		Error string
	}
)

type Exchange struct {
	Mu        sync.RWMutex
	EthClient *ethclient.Client
	btcClient *rpcclient.Client
	Users     map[int64]*user.User // orderId => User

	LimitOrders map[order.Market]map[int64][]*limit.LimitOrder // ordeers map a user to his orders
	StopOrders  map[order.Market]map[int64][]*order.StopOrder  // ordeers map a user to his orders

	ETHPrivateKey       *ecdsa.PrivateKey
	BTCAddress          string
	UstdContractAddress string

	Orderbook map[order.Market]*orderbook.Orderbook
}

func NewExchange(ustdContractAddress string, ethPrivateKey string, btcAdress string, ethClient *ethclient.Client, btcClient *rpcclient.Client) (*Exchange, error) {
	pk, err := crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		return nil, err
	}

	//--------------- orderbook ---------------
	orderbooks := make(map[order.Market]*orderbook.Orderbook)
	// orderbooks[order.MarketETH] = orderbook.NewOrderbook()
	orderbooks[order.MarketETH_Fiat] = orderbook.NewOrderbook()
	orderbooks[order.MarketETH_USDT] = orderbook.NewOrderbook()

	// orderbooks[order.MarketBTC] = orderbook.NewOrderbook()
	orderbooks[order.MarketBTC_Fiat] = orderbook.NewOrderbook()
	orderbooks[order.MarketBTC_USDT] = orderbook.NewOrderbook()

	// orderbooks[order.MarketUSDT] = orderbook.NewOrderbook()
	orderbooks[order.MarketUSDT_Fiat] = orderbook.NewOrderbook()

	//--------------- limitorders ---------------
	limitOrders := make(map[order.Market]map[int64][]*limit.LimitOrder)
	// limitOrders[order.MarketETH] = make(map[int64][]*limit.LimitOrder)
	limitOrders[order.MarketETH_Fiat] = make(map[int64][]*limit.LimitOrder)
	limitOrders[order.MarketETH_USDT] = make(map[int64][]*limit.LimitOrder)

	// limitOrders[order.MarketBTC] = make(map[int64][]*limit.LimitOrder)
	limitOrders[order.MarketBTC_Fiat] = make(map[int64][]*limit.LimitOrder)
	limitOrders[order.MarketBTC_USDT] = make(map[int64][]*limit.LimitOrder)

	// limitOrders[order.MarketUSDT] = make(map[int64][]*limit.LimitOrder)
	limitOrders[order.MarketUSDT_Fiat] = make(map[int64][]*limit.LimitOrder)

	//--------------- stoporders ---------------
	stopOrders := make(map[order.Market]map[int64][]*order.StopOrder)
	// stopOrders[order.MarketETH] = make(map[int64][]*order.StopOrder)
	stopOrders[order.MarketETH_Fiat] = make(map[int64][]*order.StopOrder)
	stopOrders[order.MarketETH_USDT] = make(map[int64][]*order.StopOrder)

	// stopOrders[order.MarketBTC] = make(map[int64][]*order.StopOrder)
	stopOrders[order.MarketBTC_Fiat] = make(map[int64][]*order.StopOrder)
	stopOrders[order.MarketBTC_USDT] = make(map[int64][]*order.StopOrder)

	// stopOrders[order.MarketUSDT] = make(map[int64][]*order.StopOrder)
	stopOrders[order.MarketUSDT_Fiat] = make(map[int64][]*order.StopOrder)

	return &Exchange{

		EthClient:           ethClient,
		btcClient:           btcClient,
		Users:               make(map[int64]*user.User),
		LimitOrders:         limitOrders,
		StopOrders:          stopOrders,
		ETHPrivateKey:       pk,
		Orderbook:           orderbooks,
		BTCAddress:          btcAdress,
		UstdContractAddress: ustdContractAddress,
	}, nil
}
