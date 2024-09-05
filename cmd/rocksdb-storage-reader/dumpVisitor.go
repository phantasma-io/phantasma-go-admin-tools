package main

import (
	"bytes"

	"github.com/linxGnu/grocksdb"
)

type DumpDataVisitor struct {
	KeyPrefix            []byte
	SubKeys1             [][]byte
	SubKeys2             [][]byte
	PanicOnUnknownSubkey bool
	Limit                uint
	limitCounter         uint
	output               *Output
}

func (v *DumpDataVisitor) Visit(it *grocksdb.Iterator) bool {
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
	result, success := DumpRow(keySlice.Data(), valueSlice.Data(), v.SubKeys1, v.SubKeys2, v.PanicOnUnknownSubkey)
	if success {
		v.output.AddAnyRecord(result)
	}

	keySlice.Free()
	valueSlice.Free()

	return true
}
