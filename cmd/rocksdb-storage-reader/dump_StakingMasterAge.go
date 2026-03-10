package main

import (
	"encoding/json"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type Visitor_StakingMasterAge struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_StakingMasterAge) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_StakingMasterAge) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var stakingMasterAges []storage.KeyValueJson

func dump_StakingMasterAge() {
	stakingMasterAges = make([]storage.KeyValueJson, 0)

	v := Visitor_StakingMasterAge{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._masterAgeMap")

	count := readLogicalCount(v.Connection, v.KeyPrefix)
	entries := collectLogicalPrefixedEntries(v.Connection, v.KeyPrefix, count)
	for _, entry := range entries {
		kr := storage.KeyValueReaderNew(entry.key)
		kr.SkipBytes(len(v.KeyPrefix))
		address := kr.ReadAddress(true)

		vr := storage.KeyValueReaderNew(entry.value)
		stakingMasterAges = append(stakingMasterAges, storage.KeyValueJson{Key: address.Text(), Value: vr.ReadTimestamp()})
	}

	row, err := json.Marshal(stakingMasterAges)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
