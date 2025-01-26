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
	Mu            sync.RWMutex
	EthClient     *ethclient.Client
	btcClient     *rpcclient.Client
	Users         map[int64]*user.User          // orderId => User
	Orders        map[int64][]*limit.LimitOrder // ordeers map a user to his orders
	ETHPrivateKey *ecdsa.PrivateKey
	BTCAddress    string
	Orderbook     map[order.Market]*orderbook.Orderbook
}

func NewExchange(ethPrivateKey string, btcAdress string, ethClient *ethclient.Client, btcClient *rpcclient.Client) (*Exchange, error) {
	pk, err := crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		return nil, err
	}

	orderbooks := make(map[order.Market]*orderbook.Orderbook)

	orderbooks[order.MarketETH] = orderbook.NewOrderbook()
	orderbooks[order.MarketBTC] = orderbook.NewOrderbook()

	return &Exchange{

		EthClient:     ethClient,
		btcClient:     btcClient,
		Users:         make(map[int64]*user.User),
		Orders:        make(map[int64][]*limit.LimitOrder),
		ETHPrivateKey: pk,
		Orderbook:     orderbooks,
		BTCAddress:    btcAdress,
	}, nil
}
