package main

import (
	"bytes"
	"math/big"
	"strings"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

var countSuffix string = "{count}"

var mapKeysAndCouns map[string]big.Int = make(map[string]big.Int)
var mapKeys []string = make([]string, 0)

type Visitor_ContractsVariables_Maps struct {
	Connection *rocksdb.Connection
	output     *Output
}

func (v *Visitor_ContractsVariables_Maps) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_ContractsVariables_Maps) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

func (v *Visitor_ContractsVariables_Maps) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	keySlice := it.Key()

	foundContract := ""
	for _, contractName := range appOpts.subKeysSlice {
		if bytes.HasPrefix(keySlice.Data(), []byte(contractName+".")) {
			foundContract = contractName
			break
		}
	}
	if foundContract == "" {
		keySlice.Free()
		return true
	}

	kr := storage.KeyValueReaderNew(keySlice.Data())
	kr.SkipBytes(len(foundContract + "."))
	keyString := kr.ReadString(false)
	if isTokenId(keyString) {
		keySlice.Free()
		return true
	}

	if strings.HasPrefix(keyString, seriePrefix) {
		kr.SkipBytes(len(seriePrefix))
		seriesPostfix := kr.ReadString(false)
		if isNumber(seriesPostfix) {
			keySlice.Free()
			return true
		}
	}

	if !strings.HasSuffix(keyString, countSuffix) {
		keySlice.Free()
		return true
	}

	// We found a map counter
	valueSlice := it.Value()
	vr := storage.KeyValueReaderNew(valueSlice.Data())
	keyString = string(keySlice.Data())
	mapKey := strings.TrimSuffix(keyString, countSuffix)
	mapKeysAndCouns[mapKey] = *vr.ReadBigInt(false)
	mapKeys = append(mapKeys, mapKey)

	keySlice.Free()
	valueSlice.Free()

	return true
}
