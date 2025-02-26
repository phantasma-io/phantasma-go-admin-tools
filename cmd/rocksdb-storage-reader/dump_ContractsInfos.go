package main

import (
	"bytes"
	"slices"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/domain/contract"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

type Visitor_ContractsInfos struct {
	Connection *rocksdb.Connection
	KeyPrefix  []byte
	output     *Output
}

func (v *Visitor_ContractsInfos) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_ContractsInfos) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

var addresses_ContractsInfos []string

func (v *Visitor_ContractsInfos) Visit(it *grocksdb.Iterator) bool {
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
	if !slices.Contains(addresses_ContractsInfos, address.Text()) {
		addresses_ContractsInfos = append(addresses_ContractsInfos, address.Text())
	}

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_ContractsInfos() {
	v := Visitor_ContractsInfos{}
	v.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)

	v.KeyPrefix = []byte(".contract.")

	v.Connection.Visit(&v)

	for _, a := range addresses_ContractsInfos {
		address, err := cryptography.FromString(a)
		if err != nil {
			panic(err)
		}

		nameValue, err := v.Connection.Get(append(append(v.KeyPrefix, address.Bytes()...), []byte(".name")...))
		if err != nil {
			panic(err)
		}

		ownerValue, err := v.Connection.Get(append(append(v.KeyPrefix, address.Bytes()...), []byte(".owner")...))
		if err != nil {
			panic(err)
		}
		vr := storage.KeyValueReaderNew(ownerValue)
		owner := vr.ReadAddress(false)

		scriptValue, err := v.Connection.Get(append(append(v.KeyPrefix, address.Bytes()...), []byte(".script")...))
		if err != nil {
			panic(err)
		}

		abiValue, err := v.Connection.Get(append(append(v.KeyPrefix, address.Bytes()...), []byte(".abi")...))
		if err != nil {
			panic(err)
		}

		v.output.AddJsonRecord(storage.ContractInfo{Address: a,
			Owner:  owner.Text(),
			Name:   string(nameValue),
			Script: scriptValue,
			ABI:    *io.Deserialize[*contract.ContractInterface_S](abiValue)})
	}

	v.Uninit()
}
