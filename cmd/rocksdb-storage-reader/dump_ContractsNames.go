package main

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

type Visitor_ContractsNames struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_ContractsNames) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_ContractsNames) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var addresses_ContractsNames []string

func (v *Visitor_ContractsNames) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	keySlice := it.Key()

	if v.KeyPrefix != nil && !bytes.HasPrefix(keySlice.Data(), v.KeyPrefix) {
		keySlice.Free()
		return true
	}

	valueSlice := it.Value()

	kr := storage.KeyValueReaderNew(keySlice.Data())
	kr.SkipBytes(len(v.KeyPrefix))
	address := kr.ReadAddress(false)
	if !slices.Contains(addresses_ContractsNames, address.Text()) {
		addresses_ContractsNames = append(addresses_ContractsNames, address.Text())
	}

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_ContractsNames() {
	v := Visitor_ContractsNames{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".contract.")

	v.Connection.Visit(&v)

	for _, a := range addresses_ContractsNames {
		address, err := cryptography.FromString(a)
		if err != nil {
			panic(err)
		}

		nameValue, err := v.Connection.Get(append(append(v.KeyPrefix, address.Bytes()...), []byte(".name")...))
		if err != nil {
			panic(err)
		}

		fmt.Println(string(nameValue))
		v.output.AddRecord(storage.KeyValue{Value: string(nameValue)})
	}

	v.Uninit()
}
