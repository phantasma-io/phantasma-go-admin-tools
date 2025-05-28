package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var cfgShowFungible bool
var cfgShowNonfungible bool
var cfgPayloadFragment string
var cfgSymbol string
var cfgEventKinds []event.EventKind
var cfgShowFailedTxes bool
var maxAttempts int = 100

func getAllAddressTransactions(address string, startingDate int64) []response.TransactionResult {
	// Calling "GetAddressTransactionCount" method to get transactions for the address
	var transactionCount int
	var err error

	for range maxAttempts {
		transactionCount, err = client.GetAddressTransactionCount(address, "main")
		if err == nil {
			break
		}
	}

	if err != nil {
		panic("GetAddressTransactionCount call failed for address '" + address + "'! Error: " + err.Error())
	}

	txs := []response.TransactionResult{}

	pageSize := 100
	pagesNumber := transactionCount / pageSize
	if transactionCount%pageSize > 0 {
		pagesNumber++
	}

	for p := 1; p <= pagesNumber; p++ {
		// Calling "GetAddressTransactions" method to get transactions for the address

		var txsResponse response.PaginatedResult[response.AddressTransactionsResult]
		for range maxAttempts {
			txsResponse, err = client.GetAddressTransactions(address, p, pageSize)
			if err == nil {
				break
			}
		}

		if err != nil {
			panic("GetAddressTransactions call failed for address '" + address + "'! Error: " + err.Error())
		}
		txs = append(txs, txsResponse.Result.Txs...)

		if startingDate > 0 && txs[len(txs)-1].Timestamp < uint(startingDate) {
			break
		}
	}

	// TODO we need to filter out all txes which have been generated in blocks after account state's block.

	return txs
}

func getCurrentAddressState(address string) response.AccountResult {
	var err error
	var account response.AccountResult
	for range maxAttempts {
		account, err = client.GetAccount(address)

		if err != nil && strings.Contains(err.Error(), "Address is invalid") {
			panic("Address is invalid: " + address)
		}

		if err == nil {
			break
		}
	}
	if err != nil {
		panic("GetAccount call failed for address '" + address + "'! Error: " + err.Error())
	}

	// TODO we need to return block number corresponding to this state
	// Otherwise we can't be 100% sure that our list of txes loaded after this call does not have tx which is not reflected in this state.
	// Block number is not yet available on API
	return account
}

func printTransactions(address string, trackAccountState, useInitialState bool, orderDirection analysis.OrderDirection, verbose, printRelatedAddresses bool) {
	if address == "" {
		panic("Address should be set")
	}

	var account response.AccountResult
	if useInitialState {
		account = getCurrentAddressState(address)
	} else {
		account.Address = address
	}

	txes := getAllAddressTransactions(address, 0)
	slices.Reverse(txes) // Reordering txes, we need them ordered from older to newer

	includedTxes := analysis.GetTransactionsByKind(txes, address, cfgSymbol, cfgPayloadFragment, cfgEventKinds, cfgShowFailedTxes)

	var rowsToPrint []string

	if trackAccountState {
		var states []analysis.AccountState
		var relatedAddresses []string
		if useInitialState {
			states, relatedAddresses = analysis.TrackAccountStateByEvents(txes, &account, analysis.Backward, verbose)
		} else {
			states, relatedAddresses = analysis.TrackAccountStateByEvents(txes, &account, analysis.Forward, verbose)
		}

		if printRelatedAddresses {
			rowsToPrint = append(rowsToPrint, relatedAddresses...)
		} else {
			rowsToPrint = analysis.DescribeTransactions(includedTxes,
				states,
				address, cfgSymbol, cfgPayloadFragment, cfgShowFungible, cfgShowNonfungible, cfgEventKinds, cfgShowFailedTxes)
		}
	} else {
		for _, t := range includedTxes {
			rowsToPrint = append(rowsToPrint, t.Hash)
		}
	}

	if orderDirection == analysis.Desc {
		slices.Reverse(rowsToPrint)
	}

	for _, r := range rowsToPrint {
		fmt.Println(r)
	}
}

func printOriginalState(address string, verbose bool) {
	if address == "" {
		panic("Address should be set")
	}

	account := getCurrentAddressState(address)

	txes := getAllAddressTransactions(address, 0)

	slices.Reverse(txes)

	analysis.TrackAccountStateByEvents(txes, &account, analysis.Backward, verbose)

	body, err := json.MarshalIndent(account, " ", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Print(string(body))
}

func printSmStates(address string, startingDate int64, verbose bool) []string {
	if address == "" {
		panic("Address should be set")
	}

	account := getCurrentAddressState(address)
	txes := getAllAddressTransactions(address, startingDate)

	slices.Reverse(txes)

	isSmNow := analysis.CheckIfSm(&account)
	states, _ := analysis.TrackAccountStateByEvents(txes, &account, analysis.Backward, verbose)

	// We process from now on to the past, so we need to revert states, latest state should be first
	slices.Reverse(states)
	eligibleMonths := analysis.DetectEligibleSm(isSmNow, states, startingDate)

	return eligibleMonths
}
