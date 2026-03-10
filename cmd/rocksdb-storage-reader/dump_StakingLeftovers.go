package main

import (
	"encoding/json"

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

func dump_StakingLeftovers() {
	stakingLeftovers = make([]storage.KeyValueJson, 0)

	v := Visitor_StakingLeftovers{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._leftoverMap")

	count := readLogicalCount(v.Connection, v.KeyPrefix)
	entries := collectLogicalPrefixedEntries(v.Connection, v.KeyPrefix, count)
	for _, entry := range entries {
		kr := storage.KeyValueReaderNew(entry.key)
		kr.SkipBytes(len(v.KeyPrefix))
		address := kr.ReadAddress(true)

		vr := storage.KeyValueReaderNew(entry.value)
		stakingLeftovers = append(stakingLeftovers, storage.KeyValueJson{Key: address.Text(), Value: vr.ReadBigInt(true).String()})
	}

	row, err := json.Marshal(stakingLeftovers)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
