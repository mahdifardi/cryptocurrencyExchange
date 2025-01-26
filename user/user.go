package user

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type User struct {
	ID            int64
	ETHPrivateKey *ecdsa.PrivateKey
	BTCAdress     string
}

func NewUser(ethPrivateKey string, btcAdress string, userId int64) *User {
	pk, err := crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:            userId,
		ETHPrivateKey: pk,
		BTCAdress:     btcAdress,
	}
}
