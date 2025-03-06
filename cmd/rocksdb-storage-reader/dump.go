package main

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/domain/contract"
	phaio "github.com/phantasma-io/phantasma-go/pkg/io"
)

func dump() {
	if appOpts.DumpNfts {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		for _, b := range appOpts.nftBalances {
			for _, id := range b.Ids {
				tokenContentBytes, err := c.Get(GetNftTokenKey(b.TokenSymbol, id))
				if err != nil {
					panic(err)
				}

				tokenContent := phaio.Deserialize[*contract.TokenContent_S](Decompress(tokenContentBytes))
				tokenContent.Symbol = b.TokenSymbol
				tokenContent.TokenID = id

				o.AddJsonRecord(tokenContent)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpSeries {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		var alreadyExported []string

		for _, b := range appOpts.nftBalances {
			for _, id := range b.Ids {
				tokenContentBytes, err := c.Get(GetNftTokenKey(b.TokenSymbol, id))
				if err != nil {
					panic(err)
				}

				tokenContent := phaio.Deserialize[*contract.TokenContent_S](Decompress(tokenContentBytes))

				if slices.Contains(alreadyExported, b.TokenSymbol+tokenContent.SeriesID) {
					continue
				}

				seriesContentBytes, err := c.Get(GetTokenSeriesKey(b.TokenSymbol, tokenContent.SeriesID))
				if err != nil {
					panic(err)
				}
				seriesContent := phaio.Deserialize[*contract.TokenSeries_S](seriesContentBytes)
				seriesContent.Symbol = b.TokenSymbol
				seriesContent.SeriesID = tokenContent.SeriesID

				o.AddJsonRecord(seriesContent)

				alreadyExported = append(alreadyExported, b.TokenSymbol+tokenContent.SeriesID)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpBlockHashes || appOpts.DumpBlocks || appOpts.DumpTokenSymbols {
		it := DumpDataMapIterator{Limit: appOpts.Limit}
		it.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

		if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
			it.KeyPrefix = []byte(Height)
		} else if appOpts.DumpTokenSymbols {
			it.KeyPrefix = []byte(TokensList)
		}

		{
			count, err := it.Connection.GetAsBigInt(storage.CountKey([]byte(it.KeyPrefix)))
			if err != nil {
				panic(err)
			}

			if appOpts.Verbose {
				fmt.Printf("Map size: %d\n", count)
			}

			var one = big.NewInt(1)
			for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
				displayKey := big.NewInt(0).Add(i, big.NewInt(1))
				it.Iterate(i, displayKey.String(), nil)
			}
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
		} else if appOpts.DumpTransactions {
			v.KeyPrefix = []byte(Txs)
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

		if appOpts.DumpTransactions {
			sort.Slice(v.output.AnyRecords, func(i, j int) bool {
				if v.output.AnyRecords[i].(storage.Tx).BlockHeightUint == v.output.AnyRecords[j].(storage.Tx).BlockHeightUint {
					fmt.Println("Block with 2 or more txes: " + v.output.AnyRecords[i].(storage.Tx).BlockHeight)
				}
				return v.output.AnyRecords[i].(storage.Tx).BlockHeightUint < v.output.AnyRecords[j].(storage.Tx).BlockHeightUint
			})
		}

		v.Uninit()
	}
}
