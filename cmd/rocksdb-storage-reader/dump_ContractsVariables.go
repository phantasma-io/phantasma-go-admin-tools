package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"unicode"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

var seriePrefix string = "serie"

type Visitor_ContractsVariables struct {
	Connection *rocksdb.Connection
	output     *Output
}

func (v *Visitor_ContractsVariables) Init(dbPath, columnFamily, outputFormat string) {
	v.Connection = rocksdb.NewConnection(dbPath, columnFamily)
	v.output = NewOutput(OutputFormatFromString(outputFormat))
}

func (v *Visitor_ContractsVariables) Uninit() {
	v.Connection.Destroy()
	v.output.Flush()
}

func isNumber(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func isTokenId(s string) bool {
	return isNumber(s) && slices.Contains(appOpts.nftTokenIds, s)
}

var contractVariables map[string]storage.ContractVariables = make(map[string]storage.ContractVariables)

func (v *Visitor_ContractsVariables) Visit(it *grocksdb.Iterator) bool {
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

	if strings.HasSuffix(keyString, countSuffix) {
		keySlice.Free()
		return true
	}

	valueSlice := it.Value()

	mapItemFound := false
	mapKey := ""
	for p, _ := range mapKeysAndCouns {
		if bytes.HasPrefix(keySlice.Data(), []byte(p)) {
			mapItemFound = true
			mapKey = p[len(foundContract+"."):len(p)]
			break
		}
	}

	contractItem, ok := contractVariables[foundContract]
	if !ok {
		contractItem = storage.ContractVariables{SingleVars: make([]storage.SingleVar, 0), MapsAndLists: make(map[string]storage.MapOfVars)}
	}

	if mapItemFound {
		kr := storage.KeyValueReaderNew(keySlice.Data())
		kr.SkipBytes(len(foundContract + "." + mapKey))

		varMap, ok := contractItem.MapsAndLists[mapKey]
		if !ok {
			count := mapKeysAndCouns[foundContract+"."+mapKey]
			varMap = storage.MapOfVars{Count: uint64(count.Int64()), Values: make([]storage.SingleVar, 0)}
		}

		key := kr.ReadBytes(false)

		v := storage.SingleVar{Key: key, Value: util.ArrayClone(valueSlice.Data())}
		varMap.Values = append(varMap.Values, v)
		contractItem.MapsAndLists[mapKey] = varMap
	} else {
		kr := storage.KeyValueReaderNew(keySlice.Data())
		kr.SkipBytes(len(foundContract + "."))

		contractItem.SingleVars = append(contractVariables[foundContract].SingleVars, storage.SingleVar{Key: kr.ReadString(false), Value: util.ArrayClone(valueSlice.Data())})
	}

	contractVariables[foundContract] = contractItem

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_ContractsVariables() {
	v1 := Visitor_ContractsVariables_Maps{}
	v1.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)
	v1.Connection.Visit(&v1)
	v1.Uninit()

	v2 := Visitor_ContractsVariables{}
	v2.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)
	v2.Connection.Visit(&v2)

	row, err := json.Marshal(contractVariables)
	if err != nil {
		panic(err)
	}
	v2.output.outputFile.WriteString(string(row))
	v2.Connection.Destroy()
}
