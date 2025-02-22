package main

import (
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/domain/token"
	phaio "github.com/phantasma-io/phantasma-go/pkg/io"
)

func dump_TokenInfo() {
	c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)
	o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

	for _, symbol := range appOpts.subKeysSlice {
		tokenInfoBytes, err := c.Get(GetTokenInfoKey(symbol))
		if err != nil {
			panic(err)
		}

		tokenInfo := phaio.Deserialize[*token.TokenInfo_S](tokenInfoBytes)

		o.AddJsonRecord(tokenInfo)
	}

	c.Destroy()
	o.Flush()
}
