package exchange

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
)

func (ex *Exchange) HandleMatches(matches []limit.Match) error {

	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.ID)
		}

		toUser, ok := ex.Users[match.Bid.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.ID)
		}

		toAddress := crypto.PubkeyToAddress(toUser.ETHPrivateKey.PublicKey)

		//exchange ffees
		// exchangePublicKey := ex.PrivateKey.Public()
		// exchangePublicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting public key to ECDSA")
		// }

		amount := big.NewInt(int64(match.SizeFilled))

		transferETH(ex.EthClient, fromUser.ETHPrivateKey, toAddress, amount)

	}

	return nil
}
