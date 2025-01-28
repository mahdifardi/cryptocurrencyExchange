package exchange

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleMatches(market order.Market, matches []limit.Match) error {

	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.ID)
		}

		toUser, ok := ex.Users[match.Bid.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.ID)
		}

		if market == order.MarketETH {

			toAddress := crypto.PubkeyToAddress(toUser.ETHPrivateKey.PublicKey)

			//exchange ffees
			// exchangePublicKey := ex.PrivateKey.Public()
			// exchangePublicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
			// if !ok {
			// 	return fmt.Errorf("error casting public key to ECDSA")
			// }

			amount := big.NewInt(int64(match.SizeFilled))

			err := transferETH(ex.EthClient, fromUser.ETHPrivateKey, toAddress, amount)
			if err != nil {
				return err
			}
		} else if market == order.MarketBTC {

			err := transferBTC(ex.btcClient, fromUser.BTCAdress, toUser.BTCAdress, match.SizeFilled)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("market does not supported")
		}

	}

	return nil
}
