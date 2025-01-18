package user

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string, userId int64) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         userId,
		PrivateKey: pk,
	}
}
