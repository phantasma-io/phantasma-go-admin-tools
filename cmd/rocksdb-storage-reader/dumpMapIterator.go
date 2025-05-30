package main

import (
	"math/big"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type DumpDataMapIterator struct {
	Connection   *rocksdb.Connection
	KeyPrefix    []byte
	SubKeys1     [][]byte
	Addresses    []string
	Limit        uint
	limitCounter uint
	output       *Output
}

func (it *DumpDataMapIterator) Init(dbPath, columnFamily, outputFormat string) {
	it.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	it.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (it *DumpDataMapIterator) Uninit() {
	it.Connection.Destroy()
	it.output.Flush()
}

func (it *DumpDataMapIterator) Iterate(index *big.Int, displayKey string, keyPrefix []byte) bool {
	if it.Connection == nil {
		panic("Connection must be set")
	}

	if it.Limit > 0 && it.limitCounter == it.Limit {
		return false
	}
	if it.Limit > 0 {
		it.limitCounter++
	}

	var key []byte
	if keyPrefix != nil {
		// Using unique key prefix for every Iterate() call.
		key = storage.ElementKey(keyPrefix, index)
	} else {
		// Using one prefix for all Iterate() calls.
		key = storage.ElementKey([]byte(it.KeyPrefix), index)
	}
	value, err := it.Connection.Get(key)
	if err != nil {
		panic(err)
	}

	result, success := DumpRow(it.Connection, key, displayKey, value, it.SubKeys1, it.Addresses, false)
	if success {
		it.output.AddRecord(result)
	}

	return true
}
