package main

import (
	"math/big"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

func dump() {
	if appOpts.BaseKey == TokensList.String() {
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

			o.AddStringRecord(storage.ReadStringWithLengthByte(v))
		}

		o.Flush()

		c.Destroy()
	} else {
		v := DumpDataVisitor{Limit: appOpts.Limit}

		if appOpts.DumpAddresses {
			v.KeyPrefix = []byte(AccountAddressMap)
		} else if appOpts.DumpBalances {
			v.KeyPrefix = []byte(Balances)
		} else if appOpts.BaseKey != "" {
			v.KeyPrefix = []byte(appOpts.BaseKey)
		}

		if appOpts.SubKeys != "" {
			subkeys1 := strings.Split(appOpts.SubKeys, ",")
			v.SubKeys1 = [][]byte{}
			for _, s := range subkeys1 {
				v.SubKeys1 = append(v.SubKeys1, []byte(s))
			}
		}
		if appOpts.SubKeys2 != "" {
			subkeys2 := strings.Split(appOpts.SubKeys2, ",")
			v.SubKeys2 = [][]byte{}
			for _, s := range subkeys2 {
				v.SubKeys2 = append(v.SubKeys2, []byte(s))
			}
		}
		v.PanicOnUnknownSubkey = appOpts.PanicOnUnknownSubkey

		v.output = NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		c.Visit(&v)
		c.Destroy()

		v.output.Flush()
	}
}
