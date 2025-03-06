package main

import (
	"bytes"
	"encoding/json"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type Visitor_StakingLeftovers struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_StakingLeftovers) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_StakingLeftovers) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var stakingLeftovers []storage.KeyValueJson

func (v *Visitor_StakingLeftovers) Visit(it *grocksdb.Iterator) bool {
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

	stakingLeftovers = append(stakingLeftovers, storage.KeyValueJson{Key: address.Text(), Value: vr.ReadBigInt(true).String()})

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_StakingLeftovers() {
	v := Visitor_StakingLeftovers{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._leftoverMap")

	v.Connection.Visit(&v)

	row, err := json.Marshal(stakingLeftovers)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
