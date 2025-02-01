package exchange

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const erc20ABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func transferUSDT(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, usdtContract common.Address, to common.Address, amount *big.Int) error {
	ctx := context.Background()

	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return err
	}

	data, err := parsedABI.Pack("transfer", to, amount)
	if err != nil {
		return err
	}

	// Estimate Gas
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		To:   &usdtContract,
		Data: data,
	})
	if err != nil {
		return err
	}

	// Create the transaction
	tx := types.NewTransaction(nonce, usdtContract, big.NewInt(0), gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	fmt.Printf("USDT Transfer Tx Hash: %s\n", signedTx.Hash().Hex())

	return nil
}
