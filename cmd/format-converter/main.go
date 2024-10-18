package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var client rpc.PhantasmaRPC
var chainTokens map[string]response.TokenResult

var appOpts struct {
	FromRocksdbToApi bool   `long:"from-rocksdb-to-api" description:"Converts data extracted from RocksDB storage into API output structures"`
	BalancesJsonRDB  string `long:"balances-json-rdb" description:"Path to JSON with account balances from RocksDB"`
	Nexus            string `long:"nexus" description:"Chain nexus to use during chain requests"`
	Verbose          bool   `short:"v" long:"verbose" description:"Verbose mode"`
}

func GetOrAddAccount(accounts []response.AccountResult, address string) (*response.AccountResult, []response.AccountResult) {
	for i, a := range accounts {
		if a.Address == address {
			return &accounts[i], accounts
		}
	}

	a := response.AccountResult{Address: address}
	accounts = append(accounts, a)

	return &accounts[len(accounts)-1], accounts
}

func main() {
	_, err := flags.Parse(&appOpts)
	if err != nil {
		panic(err)
	}

	if appOpts.Nexus == "testnet" {
		client = rpc.NewRPCTestnet()
	} else {
		client = rpc.NewRPCMainnet()
	}

	chainTokens, _ = client.GetTokensAsMap(false)

	accounts := []response.AccountResult{}

	if appOpts.FromRocksdbToApi && len(appOpts.BalancesJsonRDB) > 0 {
		byteValue, err := os.ReadFile(appOpts.BalancesJsonRDB)
		if err != nil {
			panic("ReadFile call failed! Error: " + err.Error())
		}

		var balances []storage.Balance
		json.Unmarshal(byteValue, &balances)

		for _, storageBalance := range balances {
			var a *response.AccountResult
			a, accounts = GetOrAddAccount(accounts, storageBalance.Address)
			b := a.GetTokenBalance(chainTokens[storageBalance.TokenSymbol])
			b.Chain = "main"
			b.Amount = storageBalance.Amount.String()
		}

		body, err := json.MarshalIndent(accounts, " ", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Print(string(body))
	} else {
		panic("Unsupported combination of arguments")
	}
}
