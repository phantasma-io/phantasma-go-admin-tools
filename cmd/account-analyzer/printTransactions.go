package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/console"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var transactionCount int
var transactions []response.TransactionResult
var pagination bool = true
var payloadFragment string

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

func printTransactions(address string, filterSymbol string, orderDirection analysis.OrderDirection, paginationEnabled, describeFungible, describeNonfungible bool) {
	if address == "" {
		panic("Address should be set")
	}

	transactions = getAllAddressTransactions(address)

	slices.Reverse(transactions)

	var account response.AccountResult
	account.Address = address
	perTxAccountBalances := analysis.TrackAccountStateByEvents(transactions, &account, analysis.Forward)

	transactionIndexes := makeRange(1, len(transactions))

	if paginationEnabled && orderDirection == analysis.Desc {
		slices.Reverse(transactions)
		slices.Reverse(transactionIndexes)
	}

	var pagination console.Pagination = console.Pagination{Enabled: paginationEnabled,
		ItemCount:   uint(len(transactions)),
		PageSize:    5,
		CurrentPage: 1}

	for {
		fmt.Print("\n")
		fmt.Print(
			analysis.DescribeTransactions(console.Paginate(pagination, transactions),
				console.Paginate(pagination, *perTxAccountBalances),
				console.Paginate(pagination, transactionIndexes),
				address, filterSymbol, payloadFragment, orderDirection, describeFungible, describeNonfungible))

		if pagination.Enabled {
			if !pagination.PaginationMenu() {
				break
			}
		} else {
			break
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

func printStakingTxHashes(address, filterSymbol string, orderDirection analysis.OrderDirection) {
	if address == "" {
		panic("Address should be set")
	}

	transactions = getAllAddressTransactions(address)

	fmt.Print(
		analysis.GetTransactionsByKind(transactions,
			address, filterSymbol, payloadFragment, orderDirection, event.TokenStake))
}
