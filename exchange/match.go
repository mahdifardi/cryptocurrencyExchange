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

func (ex *Exchange) HandleMatches(market order.Market, matches []limit.Match) error {

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

			fiatAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			amount := big.NewInt(int64(match.SizeFilled))

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)

			// publicKey := seller.ETHPrivateKey.Public()
			// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			// if !ok {
			// 	return fmt.Errorf("error casting public key to ECDSA")
			// }

			// sellerAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

			// sellerBalance, err := ex.EthClient.BalanceAt(ctx, sellerAddress, nil)
			// if err != nil {
			// 	return err
			// }

			gasLimit := uint64(21000) // in units

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			totalCost := new(big.Int).Add(amount,
				new(big.Int).Mul(gasPrice,
					big.NewInt(int64(gasLimit))))

			var sellerBalanceStatus bool = true
			var buyerFiatBalanceStatus bool = true

			if seller.AssetBalances[order.AsserETH].ReservedBalance.Cmp(totalCost) < 0 {
				sellerBalanceStatus = false
			}

			buyerReservedFiatBalance := buyer.AssetBalances[order.AssetFiat].ReservedBalance

			if buyerReservedFiatBalance.Cmp(fiatAmount) < 0 {
				buyerFiatBalanceStatus = false
			}

			if sellerBalanceStatus && buyerFiatBalanceStatus {
				err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, amount)
				if err != nil {
					return err
				}

				buyerETHAssetBalance := buyer.AssetBalances[order.AsserETH]
				buyerETHAssetBalance.AvailableBalance = new(big.Int).Add(buyerETHAssetBalance.AvailableBalance, amount)
				buyer.AssetBalances[order.AsserETH] = buyerETHAssetBalance

				sellerETHAssetBalance := seller.AssetBalances[order.AsserETH]
				sellerETHAssetBalance.ReservedBalance = new(big.Int).Sub(sellerETHAssetBalance.ReservedBalance, amount)
				seller.AssetBalances[order.AsserETH] = sellerETHAssetBalance

				buyerFiatAssetBalance := buyer.AssetBalances[order.AssetFiat]
				buyerFiatAssetBalance.ReservedBalance = new(big.Int).Sub(buyerFiatAssetBalance.ReservedBalance, fiatAmount)
				buyer.AssetBalances[order.AssetFiat] = buyerFiatAssetBalance

				sellerFiatAssetBalance := seller.AssetBalances[order.AssetFiat]
				sellerFiatAssetBalance.AvailableBalance = new(big.Int).Add(fiatAmount, sellerFiatAssetBalance.AvailableBalance)
				seller.AssetBalances[order.AssetFiat] = sellerFiatAssetBalance
			}
			if !sellerBalanceStatus && !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AsserETH], totalCost.String(), buyerReservedFiatBalance, totalCost)
			} else if !sellerBalanceStatus {
				return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AsserETH], totalCost.String())
			} else if !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, totalCost)

			}
		//**
		case order.MarketETH_USDT:
			// buyer quote asset (USDT) transferred to seller
			// seller base asset (ETH) transferred to buyer

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)
			sellerAddress := crypto.PubkeyToAddress(seller.ETHPrivateKey.PublicKey)

			amount := big.NewInt(int64(match.SizeFilled))

			// ----
			usdtAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
			if err != nil {
				return err
			}

			data, err := parsedABI.Pack("transfer", buyerAddress, amount)
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

			totalCostUSDT := new(big.Int).Add(amount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			var sellerEthBalanceStatus bool = true
			var buyerUSDTBalanceStatus bool = true

			buyerReservedUSDTBalance := buyer.AssetBalances[order.AsserUSDT].ReservedBalance

			if buyerReservedUSDTBalance.Cmp(totalCostUSDT) < 0 {
				buyerUSDTBalanceStatus = false
			}

			gasLimitETH := uint64(21000) // in units

			totalCostETH := new(big.Int).Add(amount,
				new(big.Int).Mul(gasPrice,
					big.NewInt(int64(gasLimitETH))))

			sellerReservedETHBalance := seller.AssetBalances[order.AsserETH].ReservedBalance
			if sellerReservedETHBalance.Cmp(totalCostETH) < 0 {
				sellerEthBalanceStatus = false
			}

			if sellerEthBalanceStatus && buyerUSDTBalanceStatus {
				err = transferETH(ex.EthClient, seller.ETHPrivateKey, buyerAddress, amount)
				if err != nil {
					return err
				}

				err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, usdtAmount)
				if err != nil {
					return err
				}

				buyerETHAssetBalance := buyer.AssetBalances[order.AsserETH]
				buyerETHAssetBalance.AvailableBalance = new(big.Int).Add(buyerETHAssetBalance.AvailableBalance, amount)
				buyer.AssetBalances[order.AsserETH] = buyerETHAssetBalance

				sellerETHAssetBalance := seller.AssetBalances[order.AsserETH]
				sellerETHAssetBalance.ReservedBalance = new(big.Int).Sub(sellerETHAssetBalance.ReservedBalance, amount)
				seller.AssetBalances[order.AsserETH] = sellerETHAssetBalance

				buyerUSDTAssetBalance := buyer.AssetBalances[order.AsserUSDT]
				buyerUSDTAssetBalance.ReservedBalance = new(big.Int).Sub(buyerUSDTAssetBalance.ReservedBalance, usdtAmount)
				buyer.AssetBalances[order.AsserUSDT] = buyerUSDTAssetBalance

				sellerUSDTAssetBalance := seller.AssetBalances[order.AsserUSDT]
				sellerUSDTAssetBalance.AvailableBalance = new(big.Int).Add(usdtAmount, sellerUSDTAssetBalance.AvailableBalance)
				seller.AssetBalances[order.AsserUSDT] = sellerUSDTAssetBalance
			}
			if !sellerEthBalanceStatus && !buyerUSDTBalanceStatus {
				return fmt.Errorf("insufficient seller ETH balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AsserETH], totalCostETH.String(), buyerReservedUSDTBalance, totalCostUSDT)
			} else if !sellerEthBalanceStatus {
				return fmt.Errorf("insufficient seller ETH balance: have %s, need %s", seller.AssetBalances[order.AsserETH], totalCostETH.String())
			} else if !buyerUSDTBalanceStatus {
				return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerReservedUSDTBalance, totalCostUSDT)

			}

		case order.MarketBTC_Fiat:

			fiatAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			amount := big.NewInt(int64(btcutil.Amount(match.SizeFilled * 1e8)))

			totalCostBTC := new(big.Int).Add(amount,
				big.NewInt(1000))

			var sellerBTCBalanceStatus bool = true
			var buyerFiatBalanceStatus bool = true

			if seller.AssetBalances[order.AsserBTC].ReservedBalance.Cmp(totalCostBTC) < 0 {
				sellerBTCBalanceStatus = false
			}

			buyerReservedFiatBalance := buyer.AssetBalances[order.AssetFiat].ReservedBalance

			if buyerReservedFiatBalance.Cmp(fiatAmount) < 0 {
				buyerFiatBalanceStatus = false
			}

			if sellerBTCBalanceStatus && buyerFiatBalanceStatus {

				err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
				if err != nil {
					return err
				}

				buyerBTCAssetBalance := buyer.AssetBalances[order.AsserBTC]
				buyerBTCAssetBalance.AvailableBalance = new(big.Int).Add(buyerBTCAssetBalance.AvailableBalance, amount)
				buyer.AssetBalances[order.AsserBTC] = buyerBTCAssetBalance

				sellerBTCAssetBalance := seller.AssetBalances[order.AsserBTC]
				sellerBTCAssetBalance.ReservedBalance = new(big.Int).Sub(sellerBTCAssetBalance.ReservedBalance, amount)
				seller.AssetBalances[order.AsserBTC] = sellerBTCAssetBalance

				buyerFiatAssetBalance := buyer.AssetBalances[order.AssetFiat]
				buyerFiatAssetBalance.ReservedBalance = new(big.Int).Sub(buyerFiatAssetBalance.ReservedBalance, fiatAmount)
				buyer.AssetBalances[order.AssetFiat] = buyerFiatAssetBalance

				sellerFiatAssetBalance := seller.AssetBalances[order.AssetFiat]
				sellerFiatAssetBalance.AvailableBalance = new(big.Int).Add(fiatAmount, sellerFiatAssetBalance.AvailableBalance)
				seller.AssetBalances[order.AssetFiat] = sellerFiatAssetBalance
			}
			if !sellerBTCBalanceStatus && !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AsserBTC], totalCostBTC.String(), buyerReservedFiatBalance, totalCostBTC)
			} else if !sellerBTCBalanceStatus {
				return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AsserBTC], totalCostBTC.String())
			} else if !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, totalCostBTC)

			}
		//**
		case order.MarketBTC_USDT:

			// buyer quote asset (USDT) transferred to seller
			// seller base asset (BTC) transferred to buyer

			buyerAddress := crypto.PubkeyToAddress(buyer.ETHPrivateKey.PublicKey)
			sellerAddress := crypto.PubkeyToAddress(seller.ETHPrivateKey.PublicKey)

			amount := big.NewInt(int64(btcutil.Amount(match.SizeFilled * 1e8)))

			// amount := big.NewInt(int64(match.SizeFilled))

			// ----
			usdtAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			gasPrice, err := ex.EthClient.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatal(err)
			}

			parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
			if err != nil {
				return err
			}

			data, err := parsedABI.Pack("transfer", buyerAddress, amount)
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

			totalCostUSDT := new(big.Int).Add(amount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			var sellerBTCBalanceStatus bool = true
			var buyerUSDTBalanceStatus bool = true

			buyerReservedUSDTBalance := buyer.AssetBalances[order.AsserUSDT].ReservedBalance

			if buyerReservedUSDTBalance.Cmp(totalCostUSDT) < 0 {
				buyerUSDTBalanceStatus = false
			}

			totalCostBTC := new(big.Int).Add(amount,
				big.NewInt(1000))

			sellerReservedBTCBalance := seller.AssetBalances[order.AsserBTC].ReservedBalance
			if sellerReservedBTCBalance.Cmp(totalCostBTC) < 0 {
				sellerBTCBalanceStatus = false
			}

			if sellerBTCBalanceStatus && buyerUSDTBalanceStatus {

				err := transferBTC(ex.btcClient, seller.BTCAdress, buyer.BTCAdress, match.SizeFilled)
				if err != nil {
					return err
				}

				err = transferUSDT(ex.EthClient, buyer.ETHPrivateKey, usdtAddress, sellerAddress, usdtAmount)
				if err != nil {
					return err
				}

				buyerBTCAssetBalance := buyer.AssetBalances[order.AsserBTC]
				buyerBTCAssetBalance.AvailableBalance = new(big.Int).Add(buyerBTCAssetBalance.AvailableBalance, amount)
				buyer.AssetBalances[order.AsserBTC] = buyerBTCAssetBalance

				sellerBTCAssetBalance := seller.AssetBalances[order.AsserBTC]
				sellerBTCAssetBalance.ReservedBalance = new(big.Int).Sub(sellerBTCAssetBalance.ReservedBalance, amount)
				seller.AssetBalances[order.AsserBTC] = sellerBTCAssetBalance

				buyerUSDTAssetBalance := buyer.AssetBalances[order.AsserUSDT]
				buyerUSDTAssetBalance.ReservedBalance = new(big.Int).Sub(buyerUSDTAssetBalance.ReservedBalance, usdtAmount)
				buyer.AssetBalances[order.AsserUSDT] = buyerUSDTAssetBalance

				sellerUSDTAssetBalance := seller.AssetBalances[order.AsserUSDT]
				sellerUSDTAssetBalance.AvailableBalance = new(big.Int).Add(usdtAmount, sellerUSDTAssetBalance.AvailableBalance)
				seller.AssetBalances[order.AsserUSDT] = sellerUSDTAssetBalance
			}
			if !sellerBTCBalanceStatus && !buyerUSDTBalanceStatus {
				return fmt.Errorf("insufficient seller BTC balance: have %s, need %s \n insufficient buyer USDT balance: have %v, need %v", seller.AssetBalances[order.AsserETH], totalCostBTC.String(), buyerReservedUSDTBalance, totalCostUSDT)
			} else if !sellerBTCBalanceStatus {
				return fmt.Errorf("insufficient seller BTC balance: have %s, need %s", seller.AssetBalances[order.AsserBTC], totalCostBTC.String())
			} else if !buyerUSDTBalanceStatus {
				return fmt.Errorf("insufficient buyer USDT balance: have %v, need %v", buyerReservedUSDTBalance, totalCostUSDT)

			}

		case order.MarketUSDT_Fiat:

			fiatAmount := new(big.Int).Mul(big.NewInt(int64(match.Price)), big.NewInt(int64(match.SizeFilled)))

			amount := big.NewInt(int64(match.SizeFilled))

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

			data, err := parsedABI.Pack("transfer", sellerAddress, amount)
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

			totalCostUSDT := new(big.Int).Add(amount, new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitUSDT))))

			var sellerUSDTBalanceStatus bool = true
			var buyerFiatBalanceStatus bool = true

			if seller.AssetBalances[order.AsserUSDT].ReservedBalance.Cmp(totalCostUSDT) < 0 {
				sellerUSDTBalanceStatus = false
			}

			buyerReservedFiatBalance := buyer.AssetBalances[order.AssetFiat].ReservedBalance

			if buyerReservedFiatBalance.Cmp(fiatAmount) < 0 {
				buyerFiatBalanceStatus = false
			}

			if sellerUSDTBalanceStatus && buyerFiatBalanceStatus {

				err = transferUSDT(ex.EthClient, seller.ETHPrivateKey, usdtAddress, buyerAddress, amount)
				if err != nil {
					return err
				}

				buyerUSDTAssetBalance := buyer.AssetBalances[order.AsserUSDT]
				buyerUSDTAssetBalance.AvailableBalance = new(big.Int).Add(buyerUSDTAssetBalance.AvailableBalance, amount)
				buyer.AssetBalances[order.AsserUSDT] = buyerUSDTAssetBalance

				sellerUSDTAssetBalance := seller.AssetBalances[order.AsserUSDT]
				sellerUSDTAssetBalance.ReservedBalance = new(big.Int).Sub(sellerUSDTAssetBalance.ReservedBalance, amount)
				seller.AssetBalances[order.AsserUSDT] = sellerUSDTAssetBalance

				buyerFiatAssetBalance := buyer.AssetBalances[order.AssetFiat]
				buyerFiatAssetBalance.ReservedBalance = new(big.Int).Sub(buyerFiatAssetBalance.ReservedBalance, fiatAmount)
				buyer.AssetBalances[order.AssetFiat] = buyerFiatAssetBalance

				sellerFiatAssetBalance := seller.AssetBalances[order.AssetFiat]
				sellerFiatAssetBalance.AvailableBalance = new(big.Int).Add(fiatAmount, sellerFiatAssetBalance.AvailableBalance)
				seller.AssetBalances[order.AssetFiat] = sellerFiatAssetBalance
			}
			if !sellerUSDTBalanceStatus && !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient seller USDT balance: have %s, need %s \n insufficient buyer Fiat balance: have %v, need %v", seller.AssetBalances[order.AsserUSDT], totalCostUSDT.String(), buyerReservedFiatBalance, fiatAmount)
			} else if !sellerUSDTBalanceStatus {
				return fmt.Errorf("insufficient seller USDT balance: have %s, need %s", seller.AssetBalances[order.AsserUSDT], totalCostUSDT.String())
			} else if !buyerFiatBalanceStatus {
				return fmt.Errorf("insufficient buyer Fiat balance: have %v, need %v", buyerReservedFiatBalance, fiatAmount)

			}

		}

		// if market == order.MarketETH {

		// 	toAddress := crypto.PubkeyToAddress(toUser.ETHPrivateKey.PublicKey)

		// 	//exchange ffees
		// 	// exchangePublicKey := ex.PrivateKey.Public()
		// 	// exchangePublicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
		// 	// if !ok {
		// 	// 	return fmt.Errorf("error casting public key to ECDSA")
		// 	// }

		// 	amount := big.NewInt(int64(match.SizeFilled))

		// 	err := transferETH(ex.EthClient, fromUser.ETHPrivateKey, toAddress, amount)
		// 	if err != nil {
		// 		return err
		// 	}
		// } else if market == order.MarketBTC {

		// 	err := transferBTC(ex.btcClient, fromUser.BTCAdress, toUser.BTCAdress, match.SizeFilled)
		// 	if err != nil {
		// 		return err
		// 	}
		// } else {
		// 	return fmt.Errorf("market does not supported")
		// }

	}

	return nil
}
