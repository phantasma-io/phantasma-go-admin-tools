package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
)

func dump() {
	if appOpts.DumpBlockHashes || appOpts.DumpBlocks || appOpts.DumpTokenSymbols {
		it := DumpDataMapIterator{Limit: appOpts.Limit}
		it.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

		if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
			it.KeyPrefix = []byte(Height)
		} else if appOpts.DumpTokenSymbols {
			it.KeyPrefix = []byte(TokensList)
		}

		count, err := it.Connection.GetAsBigInt(storage.CountKey([]byte(it.KeyPrefix)))
		if err != nil {
			panic(err)
		}

		if appOpts.Verbose {
			fmt.Printf("Map size: %d\n", count)
		}

		var one = big.NewInt(1)
		for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
			it.Iterate(i)
		}

		it.Uninit()
	} else {
		v := DumpDataVisitor{Limit: appOpts.Limit}
		v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

		if appOpts.DumpAddresses {
			v.KeyPrefix = []byte(AccountAddressMap)
		} else if appOpts.DumpBalances {
			v.KeyPrefix = []byte(Balances)
		} else if appOpts.DumpBalancesNft {
			v.KeyPrefix = []byte(Ids)
		} else if appOpts.BaseKey != "" {
			v.KeyPrefix = []byte(appOpts.BaseKey)
		}

		if len(appOpts.subKeysSlice) > 0 {
			v.SubKeys1 = [][]byte{}
			for _, s := range appOpts.subKeysSlice {
				v.SubKeys1 = append(v.SubKeys1, []byte(s))
			}
		}
		if appOpts.Addresses != "" {
			subkeys2 := strings.Split(appOpts.Addresses, ",")
			v.Addresses = []string{}
			for _, s := range subkeys2 {
				v.Addresses = append(v.Addresses, s)
			}
		}
		v.PanicOnUnknownSubkey = appOpts.PanicOnUnknownSubkey

		v.Connection.Visit(&v)

		if string(v.KeyPrefix) == Ids.String() {
			v.output.AnyRecords = storage.ConvertBalanceNonFungibleFromSingleRows(v.output.AnyRecords)
		}

		v.Uninit()
	}
}
