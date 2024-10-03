package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var transactionCount int
var transactions []response.TransactionResult
var cfgShowFungible bool
var cfgShowNonfungible bool
var cfgPayloadFragment string
var cfgSymbol string
var cfgEventKinds []event.EventKind
var cfgShowFailedTxes bool

func getAllAddressTransactions(address string) []response.TransactionResult {
	// Calling "GetAddressTransactionCount" method to get transactions for the address
	var err error
	transactionCount, err = client.GetAddressTransactionCount(address, "main")
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

func printTransactions(address string, trackAccountState bool, orderDirection analysis.OrderDirection) {
	if address == "" {
		panic("Address should be set")
	}

	txes := getAllAddressTransactions(address)
	includedTxes := analysis.GetTransactionsByKind(txes, address, cfgSymbol, cfgPayloadFragment, cfgEventKinds, cfgShowFailedTxes)

	if trackAccountState {
		slices.Reverse(txes)

		var account response.AccountResult
		account.Address = address
		perTxAccountBalances := analysis.TrackAccountStateByEvents(txes, &account, analysis.Forward)

		transactionIndexes := makeRange(1, len(txes))

		fmt.Print(
			analysis.DescribeTransactions(txes,
				includedTxes,
				*perTxAccountBalances,
				transactionIndexes,
				address, cfgSymbol, cfgPayloadFragment, orderDirection, cfgShowFungible, cfgShowNonfungible, cfgEventKinds))
	} else {
		for _, t := range includedTxes {
			fmt.Print(t.Hash)
			fmt.Println()
		}
	}
}

func printOriginalState(address string) {
	if address == "" {
		panic("Address should be set")
	}

	account := getCurrentAddressState(address)
	transactions = getAllAddressTransactions(address)

	// removing txes which are not in s.Txs list
	newLength := 0
	for index := range transactions {
		if slices.Contains(account.Txs, transactions[index].Hash) {
			transactions[newLength] = transactions[index]
			newLength++
		}
	}
	// reslice the array to remove extra index
	transactions = transactions[:newLength]

	slices.Reverse(transactions)

	perTxAccountBalances := analysis.TrackAccountStateByEventsAndCurrentState(transactions, &account, analysis.Backward)

	initialState := (*perTxAccountBalances)[0]

	body, err := json.Marshal(initialState)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(body))
}
