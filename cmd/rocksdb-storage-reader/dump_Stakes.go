package main

import (
	"bytes"
	"encoding/json"

	"github.com/linxGnu/grocksdb"
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

func (v *Visitor_Stakes) Visit(it *grocksdb.Iterator) bool {
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

	energyStake := phaio.Deserialize[*stake.EnergyStake_S](valueSlice.Data())
	stakes = append(stakes, storage.KeyValueJson{Key: address.Text(), Value: energyStake})

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_Stakes() {
	v := Visitor_Stakes{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".stake._stakeMap")

	v.Connection.Visit(&v)

	row, err := json.Marshal(stakes)
	if err != nil {
		panic(err)
	}
	v.output.outputFile.WriteString(string(row))
	v.Connection.Destroy()
}
