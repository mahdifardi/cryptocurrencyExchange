package exchange

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

func (ex *Exchange) TransferBTC(client *rpcclient.Client, fromAddress, toAddress string, amount float64) error {
	// Convert amount to satoshis (handle floating-point precision carefully)
	amountSat := btcutil.Amount(amount * 1e8)

	// Decode addresses with error wrapping
	fromAddr, err := btcutil.DecodeAddress(fromAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	toAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}

	// Fetch UTXOs with verbose error handling
	utxos, err := client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{fromAddr})
	if err != nil {
		return fmt.Errorf("failed to list UTXOs: %w", err)
	}

	// UTXO selection logic with fee estimation
	var (
		inputs       []btcjson.TransactionInput
		totalInput   btcutil.Amount
		targetAmount = amountSat + 1000 // Base fee of 1000 satoshis
	)

	for _, utxo := range utxos {
		inputs = append(inputs, btcjson.TransactionInput{
			Txid: utxo.TxID,
			Vout: utxo.Vout,
		})
		totalInput += btcutil.Amount(utxo.Amount * 1e8)

		if totalInput >= targetAmount {
			break
		}
	}

	if totalInput < amountSat {
		return fmt.Errorf("insufficient funds: need %s, available %s",
			amountSat, totalInput)
	}

	// Calculate change (handle zero-change edge case)
	change := totalInput - amountSat - 1000 // Deduct fee
	outputs := map[btcutil.Address]btcutil.Amount{
		toAddr: amountSat,
	}

	if change > 0 {
		outputs[fromAddr] = change
	}

	// Transaction construction pipeline
	rawTx, err := client.CreateRawTransaction(inputs, outputs, nil)
	if err != nil {
		return fmt.Errorf("tx creation failed: %w", err)
	}

	// Sign with wallet (ensure wallet contains private keys)
	signedTx, complete, err := client.SignRawTransactionWithWallet(rawTx)
	if err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}
	if !complete {
		return fmt.Errorf("partial signing detected")
	}

	// Broadcast with full error context
	txHash, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		return fmt.Errorf("broadcast failed: %w", err)
	}

	log.Printf("Success! TXID: %s", txHash)
	return nil
}
