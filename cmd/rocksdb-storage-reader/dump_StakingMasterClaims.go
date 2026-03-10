package main

import (
	"encoding/json"

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

func dump_StakingMasterClaims() {
	stakingMasterClaims = make([]storage.KeyValueJson, 0)

	v := Visitor_StakingMasterClaims{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._masterClaims")

	count := readLogicalCount(v.Connection, v.KeyPrefix)
	entries := collectLogicalPrefixedEntries(v.Connection, v.KeyPrefix, count)
	for _, entry := range entries {
		kr := storage.KeyValueReaderNew(entry.key)
		kr.SkipBytes(len(v.KeyPrefix))
		address := kr.ReadAddress(true)

		vr := storage.KeyValueReaderNew(entry.value)
		stakingMasterClaims = append(stakingMasterClaims, storage.KeyValueJson{Key: address.Text(), Value: vr.ReadTimestamp()})
	}

	row, err := json.Marshal(stakingMasterClaims)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
