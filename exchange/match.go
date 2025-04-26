package exchange

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mahdifardi/cryptocurrencyExchange/limit"
	"github.com/mahdifardi/cryptocurrencyExchange/order"
)

func (ex *Exchange) HandleMatches(bid bool, market order.Market, matches []limit.Match) error {

	for _, match := range matches {
		seller, ok := ex.Users[match.Ask.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.ID)
		}

		buyer, ok := ex.Users[match.Bid.UserId]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.ID)
		}
		ctx := context.Background()

		switch market {
		case order.MarketETH_Fiat:

			quoteAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			baseAmount := big.NewInt(int64(match.SizeFilled))

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)

			gasLimit := uint64(21000) // in units

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			totalCost := new(big.Int).Add(baseAmount,
				new(big.Int).Mul(gasPrice,
					big.NewInt(int64(gasLimit))))

			if bid {
				var sellerBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				sellerReservedBalabce := seller.GetReservedBalance(order.AssetETH)
				if sellerReservedBalabce.Cmp(totalCost) < 0 {
					sellerBalanceStatus = false
				}

				buyerAvailableFiatBalance := buyer.GetAvailableBalance(order.AssetFiat)

				if buyerAvailableFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerBalanceStatus && buyerFiatBalanceStatus {
					err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCost.String(), buyerAvailableFiatBalance, totalCost)
				} else if !sellerBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AssetETH], totalCost.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerAvailableFiatBalance, totalCost)

				}
			} else {
				var sellerBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				if seller.AssetBalances[order.AssetETH].AvailableBalance.Cmp(totalCost) < 0 {
					sellerBalanceStatus = false
				}

				buyerReservedFiatBalance := buyer.GetReservedBalance(order.AssetFiat)

				if buyerReservedFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerBalanceStatus && buyerFiatBalanceStatus {
					err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCost.String(), buyerReservedFiatBalance, totalCost)
				} else if !sellerBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AssetETH], totalCost.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, totalCost)

				}
			}
		//**
		case order.MarketETH_USDT:
			// buyer quote asset (USDT) transferred to seller
			// seller base asset (ETH) transferred to buyer

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)
			sellerAddress := crypto.PubkeyToAddress(seller.ETHPrivateKey.PublicKey)

			baseAmount := big.NewInt(int64(match.SizeFilled))

			// ----
			quoteAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
			if err != nil {
				return err
			}

			data, err := parsedABI.Pack("transfer", buyerAddress, baseAmount)
			if err != nil {
				return err
			}
			usdtAddress := common.HexToAddress(ex.UstdContractAddress)

			// Estimate Gas
			gasLimitUSDT, err := ex.EthClient.EstimateGas(ctx, ethereum.CallMsg{
				To:   &usdtAddress,
				Data: data,
			})
			if err != nil {
				return err
			}

			totalCostUSDT := new(big.Int).Add(baseAmount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			if bid {
				var sellerEthBalanceStatus bool = true
				var buyerUSDTBalanceStatus bool = true

				buyerAvailableUSDTBalance := buyer.GetAvailableBalance(order.AssetUSDT)

				if buyerAvailableUSDTBalance.Cmp(totalCostUSDT) < 0 {
					buyerUSDTBalanceStatus = false
				}

				gasLimitETH := uint64(21000) // in units

				totalCostETH := new(big.Int).Add(baseAmount,
					new(big.Int).Mul(gasPrice,
						big.NewInt(int64(gasLimitETH))))

				sellerReservedETHBalance := seller.GetReservedBalance(order.AssetETH)

				if sellerReservedETHBalance.Cmp(totalCostETH) < 0 {
					sellerEthBalanceStatus = false
				}

				if sellerEthBalanceStatus && buyerUSDTBalanceStatus {
					err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, quoteAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerEthBalanceStatus && !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCostETH.String(), buyerAvailableUSDTBalance, totalCostUSDT)
				} else if !sellerEthBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AssetETH], totalCostETH.String())
				} else if !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerAvailableUSDTBalance, totalCostUSDT)

				}
			} else {

				var sellerEthBalanceStatus bool = true
				var buyerUSDTBalanceStatus bool = true

				buyerReservedUSDTBalance := buyer.GetReservedBalance(order.AssetUSDT)

				if buyerReservedUSDTBalance.Cmp(totalCostUSDT) < 0 {
					buyerUSDTBalanceStatus = false
				}

				gasLimitETH := uint64(21000) // in units

				totalCostETH := new(big.Int).Add(baseAmount,
					new(big.Int).Mul(gasPrice,
						big.NewInt(int64(gasLimitETH))))

				sellerAvailableETHBalance := seller.GetAvailableBalance(order.AssetETH)
				if sellerAvailableETHBalance.Cmp(totalCostETH) < 0 {
					sellerEthBalanceStatus = false
				}

				if sellerEthBalanceStatus && buyerUSDTBalanceStatus {
					err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, quoteAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerEthBalanceStatus && !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCostETH.String(), buyerReservedUSDTBalance, totalCostUSDT)
				} else if !sellerEthBalanceStatus {
					return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AssetETH], totalCostETH.String())
				} else if !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerReservedUSDTBalance, totalCostUSDT)

				}
			}

		case order.MarketBTC_Fiat:

			quoteAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			baseAmount := big.NewInt(int64(btcutil.Amount(match.SizeFilled * 1e8)))

			totalCostBTC := new(big.Int).Add(baseAmount,
				big.NewInt(1000))

			if bid {

				var sellerBTCBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				sellerReservedBalance := seller.GetReservedBalance(order.AssetBTC)
				if sellerReservedBalance.Cmp(totalCostBTC) < 0 {
					sellerBTCBalanceStatus = false
				}

				buyerAvailableFiatBalance := buyer.GetAvailableBalance(order.AssetFiat)
				if buyerAvailableFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerBTCBalanceStatus && buyerFiatBalanceStatus {

					err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBTCBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetBTC], totalCostBTC.String(), buyerAvailableFiatBalance, totalCostBTC)
				} else if !sellerBTCBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AssetBTC], totalCostBTC.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerAvailableFiatBalance, totalCostBTC)

				}

			} else {

				var sellerBTCBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				sellerAvailableBalance := seller.GetAvailableBalance(order.AssetBTC)
				if sellerAvailableBalance.Cmp(totalCostBTC) < 0 {
					sellerBTCBalanceStatus = false
				}

				buyerReservedFiatBalance := buyer.GetReservedBalance(order.AssetFiat)

				if buyerReservedFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerBTCBalanceStatus && buyerFiatBalanceStatus {

					err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBTCBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetBTC], totalCostBTC.String(), buyerReservedFiatBalance, totalCostBTC)
				} else if !sellerBTCBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AssetBTC], totalCostBTC.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, totalCostBTC)

				}
			}
		//**
		case order.MarketBTC_USDT:

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)
			sellerAddress := crypto.PubkeyToAddress(seller.ETHPrivateKey.PublicKey)

			baseAmount := big.NewInt(int64(btcutil.Amount(match.SizeFilled * 1e8)))

			// ----
			quoteAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
			if err != nil {
				return err
			}

			data, err := parsedABI.Pack("transfer", buyerAddress, baseAmount)
			if err != nil {
				return err
			}
			usdtAddress := common.HexToAddress(ex.UstdContractAddress)

			// Estimate Gas
			gasLimitUSDT, err := ex.EthClient.EstimateGas(ctx, ethereum.CallMsg{
				To:   &usdtAddress,
				Data: data,
			})
			if err != nil {
				return err
			}

			totalCostUSDT := new(big.Int).Add(baseAmount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			if bid {

				var sellerBTCBalanceStatus bool = true
				var buyerUSDTBalanceStatus bool = true

				buyerAvailableUSDTBalance := buyer.GetAvailableBalance(order.AssetUSDT)

				if buyerAvailableUSDTBalance.Cmp(totalCostUSDT) < 0 {
					buyerUSDTBalanceStatus = false
				}

				totalCostBTC := new(big.Int).Add(baseAmount,
					big.NewInt(1000))

				sellerReservedBTCBalance := seller.GetReservedBalance(order.AssetBTC)
				if sellerReservedBTCBalance.Cmp(totalCostBTC) < 0 {
					sellerBTCBalanceStatus = false
				}

				if sellerBTCBalanceStatus && buyerUSDTBalanceStatus {

					err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
					if err != nil {
						return err
					}

					err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, quoteAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBTCBalanceStatus && !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCostBTC.String(), buyerAvailableUSDTBalance, totalCostUSDT)
				} else if !sellerBTCBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AssetBTC], totalCostBTC.String())
				} else if !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerAvailableUSDTBalance, totalCostUSDT)

				}
			} else {

				var sellerBTCBalanceStatus bool = true
				var buyerUSDTBalanceStatus bool = true

				buyerReservedUSDTBalance := buyer.GetReservedBalance(order.AssetUSDT)

				if buyerReservedUSDTBalance.Cmp(totalCostUSDT) < 0 {
					buyerUSDTBalanceStatus = false
				}

				totalCostBTC := new(big.Int).Add(baseAmount,
					big.NewInt(1000))

				sellerAvailableBTCBalance := seller.GetAvailableBalance(order.AssetBTC)
				if sellerAvailableBTCBalance.Cmp(totalCostBTC) < 0 {
					sellerBTCBalanceStatus = false
				}

				if sellerBTCBalanceStatus && buyerUSDTBalanceStatus {

					err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
					if err != nil {
						return err
					}

					err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, quoteAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerBTCBalanceStatus && !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AssetETH], totalCostBTC.String(), buyerReservedUSDTBalance, totalCostUSDT)
				} else if !sellerBTCBalanceStatus {
					return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AssetBTC], totalCostBTC.String())
				} else if !buyerUSDTBalanceStatus {
					return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerReservedUSDTBalance, totalCostUSDT)

				}

			}

		case order.MarketUSDT_Fiat:

			quoteAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			baseAmount := big.NewInt(int64(match.SizeFilled))

			sellerAddress := crypto.PubkeyToAddress(seller.ETHPrivateKey.PublicKey)

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
			if err != nil {
				return err
			}

			data, err := parsedABI.Pack("transfer", sellerAddress, baseAmount)
			if err != nil {
				return err
			}

			usdtAddress := common.HexToAddress(ex.UstdContractAddress)

			// Estimate Gas
			gasLimitUSDT, err := ex.EthClient.EstimateGas(ctx, ethereum.CallMsg{
				To:   &usdtAddress,
				Data: data,
			})
			if err != nil {
				return err
			}

			totalCostUSDT := new(big.Int).Add(baseAmount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			if bid {

				var sellerUSDTBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				if seller.AssetBalances[order.AssetUSDT].ReservedBalance.Cmp(totalCostUSDT) < 0 {
					sellerUSDTBalanceStatus = false
				}

				buyerAvailableFiatBalance := buyer.GetAvailableBalance(order.AssetFiat)

				if buyerAvailableFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerUSDTBalanceStatus && buyerFiatBalanceStatus {

					err = transferUSDT(ex.EthClient, seller.ETHPrivateKey, usdtAddress, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerUSDTBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller USDT balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetUSDT], totalCostUSDT.String(), buyerAvailableFiatBalance, quoteAmount)
				} else if !sellerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller USDT balance: have %s, need %s", seller.AssetBalances[order.AssetUSDT], totalCostUSDT.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerAvailableFiatBalance, quoteAmount)

				}
			} else {

				var sellerUSDTBalanceStatus bool = true
				var buyerFiatBalanceStatus bool = true

				sellerAvailableBalance := seller.GetAvailableBalance(order.AssetUSDT)
				if sellerAvailableBalance.Cmp(totalCostUSDT) < 0 {
					sellerUSDTBalanceStatus = false
				}

				buyerReservedFiatBalance := buyer.GetReservedBalance(order.AssetFiat)
				if buyerReservedFiatBalance.Cmp(quoteAmount) < 0 {
					buyerFiatBalanceStatus = false
				}

				if sellerUSDTBalanceStatus && buyerFiatBalanceStatus {

					err = transferUSDT(ex.EthClient, seller.ETHPrivateKey, usdtAddress, buyerAddress, baseAmount)
					if err != nil {
						return err
					}

					buyer.AddAvailableBalance(order.Asset(market.Base), baseAmount)

					seller.SubReservedBalance(order.Asset(market.Base), baseAmount)

					buyer.SubReservedBalance(order.Asset(market.Quote), quoteAmount)

					seller.AddAvailableBalance(order.Asset(market.Quote), quoteAmount)
				}
				if !sellerUSDTBalanceStatus && !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient seller USDT balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AssetUSDT], totalCostUSDT.String(), buyerReservedFiatBalance, quoteAmount)
				} else if !sellerUSDTBalanceStatus {
					return fmt.Errorf("insufficient seller USDT balance: have %s, need %s", seller.AssetBalances[order.AssetUSDT], totalCostUSDT.String())
				} else if !buyerFiatBalanceStatus {
					return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, quoteAmount)

				}

			}
		}

	}

	return nil
}
