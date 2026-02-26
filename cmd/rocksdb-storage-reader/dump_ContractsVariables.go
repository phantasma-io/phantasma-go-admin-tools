package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
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
	if !isNumber(s) {
		return false
	}

	_, ok := appOpts.nftTokenIdSet[s]
	return ok
}

var contractVariables map[string]storage.ContractVariables = make(map[string]storage.ContractVariables)

func matchContractNamespace(key []byte) (string, string, bool) {
	for _, contractName := range appOpts.subKeysSlice {
		// Most namespaces use "<name>.<field>", but token metadata is stored as
		// ".token:<symbol>". We support both separators to avoid silent drops.
		if bytes.HasPrefix(key, []byte(contractName+".")) {
			return contractName, ".", true
		}
		if bytes.HasPrefix(key, []byte(contractName+":")) {
			return contractName, ":", true
		}
	}

	return "", "", false
}

func (v *Visitor_ContractsVariables) Visit(it *grocksdb.Iterator) bool {
	if v.Connection == nil {
		panic("Connection must be set")
	}

	keySlice := it.Key()

	foundContract, keySeparator, found := matchContractNamespace(keySlice.Data())
	if !found {
		keySlice.Free()
		return true
	}

	kr := storage.KeyValueReaderNew(keySlice.Data())
	kr.SkipBytes(len(foundContract + keySeparator))
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
	for _, p := range mapKeys {
		if bytes.HasPrefix(keySlice.Data(), []byte(p)) {
			mapItemFound = true
			mapKey = p[len(foundContract+keySeparator):]
			break
		}
	}

	contractItem, ok := contractVariables[foundContract]
	if !ok {
		contractItem = storage.ContractVariables{SingleVars: make([]storage.SingleVar, 0), MapsAndLists: make(map[string]storage.MapOfVars)}
	}

	if mapItemFound {
		kr := storage.KeyValueReaderNew(keySlice.Data())
		kr.SkipBytes(len(foundContract + keySeparator + mapKey))

		varMap, ok := contractItem.MapsAndLists[mapKey]
		if !ok {
			count := mapKeysAndCouns[foundContract+keySeparator+mapKey]
			varMap = storage.MapOfVars{Count: uint64(count.Int64()), Values: make([]storage.SingleVar, 0)}
		}

		key := kr.ReadBytes(false)

		v := storage.SingleVar{Key: key, Value: util.ArrayClone(valueSlice.Data())}
		varMap.Values = append(varMap.Values, v)
		contractItem.MapsAndLists[mapKey] = varMap
	} else {
		kr := storage.KeyValueReaderNew(keySlice.Data())
		kr.SkipBytes(len(foundContract + keySeparator))

		contractItem.SingleVars = append(contractVariables[foundContract].SingleVars, storage.SingleVar{Key: kr.ReadString(false), Value: util.ArrayClone(valueSlice.Data())})
	}

	contractVariables[foundContract] = contractItem

	keySlice.Free()
	valueSlice.Free()

	return true
}

func dump_ContractsVariables() {
	// Keep state local to the current run. These accumulators are package-level
	// variables and must be reset before each dump to avoid stale carry-over if
	// multiple dump modes are ever invoked in a single process.
	mapKeysAndCouns = make(map[string]big.Int)
	mapKeys = make([]string, 0)
	contractVariables = make(map[string]storage.ContractVariables)

	v1 := Visitor_ContractsVariables_Maps{}
	v1.Init(appOpts.DbPath, appOpts.ColumnFamily, appOpts.OutputFormat)
	v1.Connection.Visit(&v1)
	v1.Uninit()

	// Very important to sort map keys here in descending order
	// to detect "timesDistributed" map first and "times" map next, not vise versa
	// Otherwise maps' content will be stored incorrectly
	slices.Sort(mapKeys)
	slices.Reverse(mapKeys)
	for _, mName := range mapKeys {
		mCount, ok := mapKeysAndCouns[mName]
		if !ok {
			panic("Map not found: " + mName)
		}
		if appOpts.Verbose {
			fmt.Printf("Map or list: %s - %d\n", mName, mCount.Int64())
		}
	}

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
