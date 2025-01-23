package main

import (
	"bytes"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type DumpDataVisitor struct {
	Connection           *rocksdb.Connection
	KeyPrefix            []byte
	SubKeys1             [][]byte
	Addresses            []string
	PanicOnUnknownSubkey bool
	Limit                uint
	limitCounter         uint
	output               *Output
}

func (v *DumpDataVisitor) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *DumpDataVisitor) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

func (v *DumpDataVisitor) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	if v.Limit > 0 && v.limitCounter == v.Limit {
		return false
	}
	if v.Limit > 0 {
		v.limitCounter++
	}

	keySlice := it.Key()

	if v.KeyPrefix != nil && !bytes.HasPrefix(keySlice.Data(), v.KeyPrefix) {
		keySlice.Free()
		return true
	}

	valueSlice := it.Value()

	result, success := DumpRow(v.Connection, keySlice.Data(), "", valueSlice.Data(), v.SubKeys1, v.Addresses, v.PanicOnUnknownSubkey)
	if success {
		v.output.AddRecord(result)
	}

	keySlice.Free()
	valueSlice.Free()

	return true
}
