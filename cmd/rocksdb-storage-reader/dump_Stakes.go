package main

import (
	"encoding/json"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/domain/stake"
	phaio "github.com/phantasma-io/phantasma-go/pkg/io"
)

type Visitor_Stakes struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_Stakes) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_Stakes) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var stakes []storage.KeyValueJson

func dump_Stakes() {
	stakes = make([]storage.KeyValueJson, 0)

	v := Visitor_Stakes{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._stakeMap")

	count := readLogicalCount(v.Connection, v.KeyPrefix)
	entries := collectLogicalPrefixedEntries(v.Connection, v.KeyPrefix, count)
	for _, entry := range entries {
		kr := storage.KeyValueReaderNew(entry.key)
		kr.SkipBytes(len(v.KeyPrefix))
		address := kr.ReadAddress(true)

		energyStake := phaio.Deserialize[*stake.EnergyStake_S](entry.value)
		stakes = append(stakes, storage.KeyValueJson{Key: address.Text(), Value: energyStake})
	}

	row, err := json.Marshal(stakes)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
