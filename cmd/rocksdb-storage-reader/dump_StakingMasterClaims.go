package main

import (
	"bytes"
	"encoding/json"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type Visitor_StakingMasterClaims struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_StakingMasterClaims) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_StakingMasterClaims) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var stakingMasterClaims []storage.KeyValueJson

func (v *Visitor_StakingMasterClaims) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	keySlice := it.Key()

	if v.KeyPrefix != nil && !bytes.HasPrefix(keySlice.Data(), v.KeyPrefix) {
		keySlice.Free()
		return true
	}
	if bytes.HasSuffix(keySlice.Data(), []byte(countSuffix)) {
		keySlice.Free()
		return true
	}

	valueSlice := it.Value()

	kr := storage.KeyValueReaderNew(keySlice.Data())
	kr.SkipBytes(len(v.KeyPrefix))
	address := kr.ReadAddress(true)

	vr := storage.KeyValueReaderNew(valueSlice.Data())

	stakingMasterClaims = append(stakingMasterClaims, storage.KeyValueJson{Key: address.Text(), Value: vr.ReadTimestamp()})

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_StakingMasterClaims() {
	v := Visitor_StakingMasterClaims{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._masterClaims")

	v.Connection.Visit(&v)

	row, err := json.Marshal(stakingMasterClaims)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
