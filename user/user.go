package user

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

type AssetBalance struct {
	AvailableBalance *big.Int
	ReservedBalance  *big.Int
}
type User struct {
	Mu            sync.RWMutex
	ID            int64
	ETHPrivateKey *ecdsa.PrivateKey
	BTCAdress     string
	AssetBalances map[order.Asset]AssetBalance
}

func NewUser(ethPrivateKey string, btcAdress string, userId int64) *User {
	pk, err := crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		panic(err)
	}

	assetBalances := make(map[order.Asset]AssetBalance)
	assetBalances[order.AssetFiat] = AssetBalance{
		AvailableBalance: big.NewInt(1_000_000_000_000),
		ReservedBalance:  big.NewInt(0),
	}

	assetBalances[order.AsserBTC] = AssetBalance{
		AvailableBalance: big.NewInt(1_000_000),
		ReservedBalance:  big.NewInt(0),
	}
	assetBalances[order.AsserETH] = AssetBalance{
		AvailableBalance: big.NewInt(1_000_000_000_000_000_000),
		ReservedBalance:  big.NewInt(0),
	}
	assetBalances[order.AsserUSDT] = AssetBalance{
		AvailableBalance: big.NewInt(1_000_000),
		ReservedBalance:  big.NewInt(0),
	}

	return &User{
		Mu:            sync.RWMutex{},
		ID:            userId,
		ETHPrivateKey: pk,
		BTCAdress:     btcAdress,
		AssetBalances: assetBalances,
	}
}

func (u *User) ReserveBalance(market order.Market, amount *big.Int, bid bool) error {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	if bid {
		userQuoteAvailableBalance := u.AssetBalances[order.Asset(market.Quote)].AvailableBalance
		if userQuoteAvailableBalance.Cmp(amount) < 0 {
			return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Quote, userQuoteAvailableBalance, amount)
		} else {
			userQuoteBalance := u.AssetBalances[order.Asset(market.Quote)]
			userQuoteBalance.AvailableBalance = new(big.Int).Sub(userQuoteBalance.AvailableBalance, amount)
			userQuoteBalance.ReservedBalance = new(big.Int).Add(userQuoteBalance.ReservedBalance, amount)
			u.AssetBalances[order.Asset(market.Quote)] = userQuoteBalance
		}
	} else {
		userBaseBalance := u.AssetBalances[order.Asset(market.Base)].AvailableBalance
		if userBaseBalance.Cmp(amount) < 0 {
			return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Quote, userBaseBalance, amount)
		} else {
			userBaseBalance := u.AssetBalances[order.Asset(market.Base)]
			userBaseBalance.AvailableBalance = new(big.Int).Sub(userBaseBalance.AvailableBalance, amount)
			userBaseBalance.ReservedBalance = new(big.Int).Add(userBaseBalance.ReservedBalance, amount)
			u.AssetBalances[order.Asset(market.Base)] = userBaseBalance
		}
	}

	return nil
}

func (u *User) GetAvailableBalance(asset order.Asset) *big.Int {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	return u.AssetBalances[asset].AvailableBalance
}

func (u *User) GetReservedBalance(asset order.Asset) *big.Int {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	return u.AssetBalances[asset].ReservedBalance
}

func (u *User) getAssetBalance(asset order.Asset) AssetBalance {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	return u.AssetBalances[asset]
}

func (u *User) AddAvailableBalance(asset order.Asset, amount *big.Int) error {
	// u.Mu.Lock()
	// defer u.Mu.Unlock()
	userAssetBalance := u.getAssetBalance(asset)
	// u.Mu.Unlock()
	userAssetBalance.AvailableBalance = new(big.Int).Add(u.GetAvailableBalance(asset), amount)
	u.Mu.Lock()
	defer u.Mu.Unlock()
	u.AssetBalances[asset] = userAssetBalance

	return nil
}

func (u *User) SubReservedBalance(asset order.Asset, amount *big.Int) error {
	// u.Mu.Lock()
	// defer u.Mu.Unlock()
	userAssetBalance := u.getAssetBalance(asset)
	userAssetBalance.ReservedBalance = new(big.Int).Sub(u.GetReservedBalance(asset), amount)
	u.Mu.Lock()
	defer u.Mu.Unlock()
	u.AssetBalances[asset] = userAssetBalance

	return nil
}
