package main

import (
	"fmt"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/console"
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
	perTxAccountBalances := analysis.TrackAccountStateByEvents(transactions, address, &account, analysis.Forward)

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
