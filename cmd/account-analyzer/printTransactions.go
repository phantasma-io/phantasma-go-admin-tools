package main

import (
	"encoding/json"
	"fmt"
	"slices"

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

func getAllAddressTransactions(address string, includeTxes []string) []response.TransactionResult {
	// Calling "GetAddressTransactionCount" method to get transactions for the address
	transactionCount, err := client.GetAddressTransactionCount(address, "main")
	if err != nil {
		panic("GetAddressTransactionCount call failed! Error: " + err.Error())
	}

	txs := []response.TransactionResult{}

	pageSize := 100
	pagesNumber := transactionCount / pageSize
	if transactionCount%pageSize > 0 {
		pagesNumber++
	}

	for p := 1; p <= pagesNumber; p++ {
		// Calling "GetAddressTransactions" method to get transactions for the address
		txsResponse, err := client.GetAddressTransactions(address, p, pageSize)
		if err != nil {
			panic("GetAddressTransactions call failed! Error: " + err.Error())
		}
		txs = append(txs, txsResponse.Result.Txs...)
	}

	if includeTxes != nil && len(includeTxes) != 0 {
		// removing txes which are not in s.Txs list
		newLength := 0
		for index := range txs {
			if slices.Contains(includeTxes, txs[index].Hash) {
				txs[newLength] = txs[index]
				newLength++
			}
		}
		// reslice the array to remove extra index
		txs = txs[:newLength]
	}

	return txs
}

func getCurrentAddressState(address string) response.AccountResult {
	// Calling "GetAddressTransactionCount" method to get transactions for the address
	var err error
	account, err := client.GetAccountEx(address)
	if err != nil {
		panic("GetAccountEx call failed! Error: " + err.Error())
	}

	return account
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func printTransactions(address string, trackAccountState, useInitialState bool, orderDirection analysis.OrderDirection) {
	if address == "" {
		panic("Address should be set")
	}

	var account response.AccountResult
	if useInitialState {
		account = getCurrentAddressState(address)
	} else {
		account.Address = address
	}

	txes := getAllAddressTransactions(address, account.Txs)
	includedTxes := analysis.GetTransactionsByKind(txes, address, cfgSymbol, cfgPayloadFragment, cfgEventKinds, cfgShowFailedTxes)

	var rowsToPrint []string

	if trackAccountState {
		slices.Reverse(txes)

		var perTxAccountBalances *[]analysis.AccountState
		if useInitialState {
			perTxAccountBalances = analysis.TrackAccountStateByEventsAndCurrentState(txes, &account, analysis.Backward)
			*perTxAccountBalances = (*perTxAccountBalances)[1:]
		} else {
			perTxAccountBalances = analysis.TrackAccountStateByEvents(txes, &account, analysis.Forward)
		}

		transactionIndexes := makeRange(1, len(txes))

		rowsToPrint = analysis.DescribeTransactions(includedTxes,
			*perTxAccountBalances,
			transactionIndexes,
			address, cfgSymbol, cfgPayloadFragment, cfgShowFungible, cfgShowNonfungible, cfgEventKinds, cfgShowFailedTxes)
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

func printOriginalState(address string) {
	if address == "" {
		panic("Address should be set")
	}

	account := getCurrentAddressState(address)
	txes := getAllAddressTransactions(address, account.Txs)

	slices.Reverse(txes)

	perTxAccountBalances := analysis.TrackAccountStateByEventsAndCurrentState(txes, &account, analysis.Backward)

	initialState := (*perTxAccountBalances)[0]

	body, err := json.MarshalIndent(initialState.State, " ", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Print(string(body))
}
