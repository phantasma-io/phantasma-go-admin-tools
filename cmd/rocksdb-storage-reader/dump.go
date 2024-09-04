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

			o.AddAnyRecord(storage.ReadStringWithLengthByte(v))
		}

		o.Flush()

		c.Destroy()
	} else {
		v := DumpDataVisitor{Limit: appOpts.Limit}
		if appOpts.BaseKey != "" {
			v.KeyPrefix = []byte(appOpts.BaseKey)
		}
		if appOpts.SubKeys != "" {
			subkeys := strings.Split(appOpts.SubKeys, ",")
			v.SubKeys = [][]byte{}
			for _, s := range subkeys {
				v.SubKeys = append(v.SubKeys, []byte(s))
			}
			v.PanicOnUnknownSubkey = appOpts.PanicOnUnknownSubkey

			v.KeyPrefix = []byte(appOpts.BaseKey)
		}

		v.output = NewOutput(OutputFormatFromString(appOpts.OutputFormat))

		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
		c.Visit(&v)
		c.Destroy()

		v.output.Flush()
	}
}
