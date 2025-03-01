package user

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

type AssetBalance struct {
	AvailableBalance *big.Int
	ReservedBalance  *big.Int
}
type User struct {
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
		AvailableBalance: big.NewInt(1_000_000),
		ReservedBalance:  big.NewInt(0),
	}

	assetBalances[order.AsserBTC] = AssetBalance{
		AvailableBalance: big.NewInt(0),
		ReservedBalance:  big.NewInt(0),
	}
	assetBalances[order.AsserETH] = AssetBalance{
		AvailableBalance: big.NewInt(0),
		ReservedBalance:  big.NewInt(0),
	}
	assetBalances[order.AsserUSDT] = AssetBalance{
		AvailableBalance: big.NewInt(0),
		ReservedBalance:  big.NewInt(0),
	}

	return &User{
		ID:            userId,
		ETHPrivateKey: pk,
		BTCAdress:     btcAdress,
		AssetBalances: assetBalances,
	}
}
