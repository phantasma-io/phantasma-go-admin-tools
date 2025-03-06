package main

import (
	"bytes"
	"strings"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

var addresses_StakingClaims []string = make([]string, 0)

type Visitor_StakingClaims_Maps struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_StakingClaims_Maps) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_StakingClaims_Maps) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

func (v *Visitor_StakingClaims_Maps) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	keySlice := it.Key()

	if v.KeyPrefix != nil && !bytes.HasPrefix(keySlice.Data(), v.KeyPrefix) {
		keySlice.Free()
		return true
	}
	if !bytes.HasSuffix(keySlice.Data(), []byte(countSuffix)) {
		keySlice.Free()
		return true
	}

	kr := storage.KeyValueReaderNew(keySlice.Data())
	kr.SkipBytes(len(v.KeyPrefix))
	address := strings.TrimSuffix(kr.ReadString(false), countSuffix)

	addresses_StakingClaims = append(addresses_StakingClaims, address)
	keySlice.Free()

	return true
}
