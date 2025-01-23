package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

func dump() {
	if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
		it := DumpDataMapIterator{Limit: appOpts.Limit}

		if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
			it.KeyPrefix = []byte(Height)
		}

		it.Connection = rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		it.output = NewOutput(OutputFormatFromString(appOpts.OutputFormat))

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

		it.Connection.Destroy()

		it.output.Flush()
	} else if appOpts.BaseKey == TokensList.String() {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)

		count, err := c.GetAsBigInt(storage.CountKey([]byte(appOpts.BaseKey)))
		if err != nil {
			panic(err)
		}

		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		var one = big.NewInt(1)
		for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
			v, err := c.Get(storage.ElementKey([]byte(appOpts.BaseKey), i))
			if err != nil {
				panic(err)
			}

			vr := storage.KeyValueReaderNew(v)
			o.AddStringRecord(vr.ReadString(true))
		}

		o.Flush()

		c.Destroy()
	} else {
		v := DumpDataVisitor{Limit: appOpts.Limit}

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

		v.output = NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		v.Connection = rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		v.Connection.Visit(&v)
		v.Connection.Destroy()

		if string(v.KeyPrefix) == Ids.String() {
			v.output.AnyRecords = storage.ConvertBalanceNonFungibleFromSingleRows(v.output.AnyRecords)
		}

		v.output.Flush()
	}
}
