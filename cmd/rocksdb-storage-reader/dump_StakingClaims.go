package main

import (
	"math/big"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
)

func dump_StakingClaims() {
	v0 := Visitor_StakingClaims_Maps{}
	v0.KeyPrefix = []byte(".stake._claimMap.")
	v0.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)
	v0.Connection.Visit(&v0)
	v0.Uninit()

	it := DumpDataMapIterator{Limit: appOpts.Limit}
	it.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)
	it.KeyPrefix = v0.KeyPrefix

	for _, address := range addresses_StakingClaims {
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

	it.Uninit()
}
