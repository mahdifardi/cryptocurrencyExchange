package exchange

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

func TransferBTC(client *rpcclient.Client, fromAddress, toAddress string, amount float64) error {
	// convert the amount of transaction to statoshi
	amountSatoshis := int64(amount * 1e8)

	// convert fromAddress to btcutil.Address
	fromAddressDecoded, err := btcutil.DecodeAddress(fromAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return fmt.Errorf("failed to decode address")
	}

	// get all utxos for fromAddress that want to send btc to toAddress
	utxos, err := client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{fromAddressDecoded})
	if err != nil {
		return fmt.Errorf("failed to fetch UTXOs: %w", err)
	}

	// a list of utxos transactions of fromAddress
	var inputs []btcjson.TransactionInput
	// total utxos of fromAddress that we want to send amountSatoshis to toAddress
	var inputAmount int64

	// Gather enough UTXOs to cover the amount + fees
	for _, utxo := range utxos {
		input := btcjson.TransactionInput{
			Txid: utxo.TxID,
			Vout: utxo.Vout,
		}
		inputs = append(inputs, input)
		inputAmount += int64(utxo.Amount * 1e8)
		if inputAmount >= amountSatoshis+1000 { // Include a fee buffer
			break
		}
	}

	// check for sufficient balance
	if inputAmount < amountSatoshis {
		return fmt.Errorf("insufficient balance: need %d satoshis, but have %d satoshis", amountSatoshis, inputAmount)
	}

	// Convert addresses to btcutil.Address type
	toAddressDecoded, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}

	// Prepare outputs (recipient and change)
	change := inputAmount - amountSatoshis - 1000 // Subtract a fee of 1000 satoshis (0.00001 BTC)
	outputs := map[btcutil.Address]btcutil.Amount{
		toAddressDecoded: btcutil.Amount(amountSatoshis),
	}

	// if there is any change from transaction, it should return to fromAddress
	if change > 0 {
		outputs[fromAddressDecoded] = btcutil.Amount(change)
	}

	// Create the raw transaction
	rawTx, err := client.CreateRawTransaction(inputs, outputs, nil)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	// Sign the raw transaction
	signedTx, _, err := client.SignRawTransaction(rawTx)
	if err != nil {
		return fmt.Errorf("failed to sign raw transaction: %w", err)
	}

	// Broadcast the signed transaction
	txHash, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	log.Printf("Transaction broadcasted! TXID: %s", txHash.String())
	return nil

}
