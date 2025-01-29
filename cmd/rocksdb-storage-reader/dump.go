package main

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"math/big"
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

				tokenContent := phaio.Deserialize[*contract.TokenContent](Decompress(tokenContentBytes))
				tokenContent.Symbol = b.TokenSymbol
				tokenContent.TokenID = BigIntFromString(id)

				o.AddJsonRecord(tokenContent)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpSeries {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		for _, b := range appOpts.nftBalances {
			for _, id := range b.Ids {
				tokenContentBytes, err := c.Get(GetNftTokenKey(b.TokenSymbol, id))
				if err != nil {
					panic(err)
				}

				tokenContent := phaio.Deserialize[*contract.TokenContent](Decompress(tokenContentBytes))

				seriesContentBytes, err := c.Get(GetTokenSeriesKey(b.TokenSymbol, tokenContent.SeriesID))
				if err != nil {
					panic(err)
				}
				seriesContent := phaio.Deserialize[*contract.TokenSeries](seriesContentBytes)

				o.AddJsonRecord(seriesContent)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpNfts {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		for _, b := range appOpts.nftBalances {
			for _, id := range b.Ids {
				key := []byte(b.TokenSymbol + "." + id)
				tokenContentBytes, err := c.Get(key)
				if err != nil {
					panic(err)
				}

				flateReader := flate.NewReader(bytes.NewReader(tokenContentBytes))
				bytesDecompressed, err := io.ReadAll(flateReader)
				if err != nil {
					panic(err)
				}
				tokenContentBytes = bytesDecompressed

				tokenContent := phaio.Deserialize[*contract.TokenContent](tokenContentBytes)
				tokenContent.Symbol = b.TokenSymbol
				tId := big.NewInt(0)
				tId.SetString(id, 10)
				tokenContent.TokenID = tId

				o.AddJsonRecord(tokenContent)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpStakes || appOpts.DumpStakingLeftovers || appOpts.DumpStakingMasterAge || appOpts.DumpStakingMasterClaims {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		var keyPrefix []byte
		if appOpts.DumpStakes {
			keyPrefix = []byte(".stake._stakeMap")
		} else if appOpts.DumpStakingLeftovers {
			keyPrefix = []byte(".stake._leftoverMap")
		} else if appOpts.DumpStakingMasterAge {
			keyPrefix = []byte(".stake._masterAgeMap")
		} else if appOpts.DumpStakingMasterClaims {
			keyPrefix = []byte(".stake._masterClaims")
		}

		if appOpts.Verbose {
			fmt.Printf("Addresses count: %d\n", len(appOpts.subKeysSlice))
		}
		for _, address := range appOpts.subKeysSlice {
			// Go through all addresses

			key := storage.KeyBuilderNew().SetBytes(keyPrefix).AppendAddressPrefixedAsString(address).Build()

			value, err := c.Get(key)
			if err != nil {
				panic(err)
			}
			if value == nil || len(value) == 0 {
				continue
			}
			result, success := DumpRow(c, key, address, value, nil, nil, false)
			if success {
				o.AddRecord(result)
			}
		}

		c.Destroy()
		o.Flush()
	} else if appOpts.DumpBlockHashes || appOpts.DumpBlocks || appOpts.DumpTokenSymbols || appOpts.DumpStakingClaims {
		it := DumpDataMapIterator{Limit: appOpts.Limit}
		it.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

		if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
			it.KeyPrefix = []byte(Height)
		} else if appOpts.DumpTokenSymbols {
			it.KeyPrefix = []byte(TokensList)
		} else if appOpts.DumpStakingClaims {
			it.KeyPrefix = []byte(".stake._claimMap.")
		}

		if appOpts.DumpStakingClaims {
			if appOpts.Verbose {
				fmt.Printf("Addresses count: %d\n", len(appOpts.subKeysSlice))
			}
			for _, address := range appOpts.subKeysSlice {
				// Go through all addresses

				keyPrefix := storage.KeyBuilderNew().SetBytes([]byte(it.KeyPrefix)).AppendString(address).Build()

				count, err := it.Connection.GetAsBigInt(storage.CountKey(keyPrefix))
				if err != nil {
					panic(err)
				}

				var one = big.NewInt(1)
				for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
					it.Iterate(i, address, keyPrefix)
				}
			}
		} else {
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

		v.Uninit()
	}
}
