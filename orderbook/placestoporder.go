package orderbook

import (
	"math/big"

	"github.com/mahdifardi/cryptocurrencyExchange/order"
	"github.com/mahdifardi/cryptocurrencyExchange/user"
)

func (ob *Orderbook) PlaceStopOrder(o *order.StopOrder, market order.Market, user *user.User) error {

	if o.Bid {
		quoteAmount := new(big.Int).Mul(big.NewInt(int64(o.Price)), big.NewInt(int64(o.Size)))
		err := user.ReserveBalance(market, quoteAmount, o.Bid)
		if err != nil {
			return err
		}
		// userQuoteAvailableBalance := user.AssetBalances[order.Asset(market.Quote)].AvailableBalance
		// if userQuoteAvailableBalance.Cmp(quoteAmount) < 0 {
		// 	return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Quote, userQuoteAvailableBalance, quoteAmount)
		// } else {
		// 	userQuoteBalance := user.AssetBalances[order.Asset(market.Quote)]
		// 	userQuoteBalance.AvailableBalance = new(big.Int).Sub(userQuoteBalance.AvailableBalance, quoteAmount)
		// 	userQuoteBalance.ReservedBalance = new(big.Int).Add(userQuoteBalance.ReservedBalance, quoteAmount)
		// 	user.AssetBalances[order.Asset(market.Quote)] = userQuoteBalance
		// }
	} else {
		BaseAmount := big.NewInt(int64(o.Size))
		err := user.ReserveBalance(market, BaseAmount, o.Bid)
		if err != nil {
			return err
		}
		// userBaseBalance := user.AssetBalances[order.Asset(market.Base)].AvailableBalance
		// if userBaseBalance.Cmp(BaseAmount) < 0 {
		// 	return fmt.Errorf("insufficient user %s balance: have %s, need %s", market.Base, userBaseBalance, BaseAmount)
		// } else {
		// 	userBaseBalance := user.AssetBalances[order.Asset(market.Base)]
		// 	userBaseBalance.AvailableBalance = new(big.Int).Sub(userBaseBalance.AvailableBalance, BaseAmount)
		// 	userBaseBalance.ReservedBalance = new(big.Int).Add(userBaseBalance.ReservedBalance, BaseAmount)
		// 	user.AssetBalances[order.Asset(market.Base)] = userBaseBalance
		// }
	}

	if o.Limit {
		ob.stopLimitOrders = append(ob.stopLimitOrders, o)
	} else {
		ob.stopMarketOrders = append(ob.stopMarketOrders, o)
	}

	return nil
}
