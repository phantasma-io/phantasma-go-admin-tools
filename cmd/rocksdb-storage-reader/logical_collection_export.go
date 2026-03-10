package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type rawKv struct {
	key   []byte
	value []byte
}

func singleVarKeyBytes(v storage.SingleVar) []byte {
	key, ok := v.Key.([]byte)
	if !ok {
		panic(fmt.Sprintf("unexpected SingleVar.Key type %T", v.Key))
	}
	return key
}

func isListElementKey(key []byte) bool {
	return len(key) >= 3 && key[0] == '<' && key[len(key)-1] == '>'
}

func hasListZeroKey(values []storage.SingleVar) bool {
	zeroKey := storage.ElementKey([]byte{}, big.NewInt(0))
	for _, v := range values {
		if bytes.Equal(singleVarKeyBytes(v), zeroKey) {
			return true
		}
	}
	return false
}

func looksLikeListCollection(values []storage.SingleVar) bool {
	if len(values) == 0 {
		return false
	}

	if hasListZeroKey(values) {
		return true
	}

	listKeys := 0
	otherKeys := 0
	for _, v := range values {
		if isListElementKey(singleVarKeyBytes(v)) {
			listKeys++
		} else {
			otherKeys++
		}
	}

	return listKeys > 0 && otherKeys == 0
}

func normalizeLogicalListValues(tableName string, count uint64, values []storage.SingleVar) ([]storage.SingleVar, error) {
	if count == 0 {
		return []storage.SingleVar{}, nil
	}

	byKey := make(map[string]storage.SingleVar, len(values))
	for _, v := range values {
		key := singleVarKeyBytes(v)
		byKey[string(key)] = v
	}

	result := make([]storage.SingleVar, 0, count)
	for i := uint64(0); i < count; i++ {
		expected := storage.ElementKey([]byte{}, big.NewInt(int64(i)))
		v, ok := byKey[string(expected)]
		if !ok {
			return nil, fmt.Errorf("list table %q missing live element at index %d", tableName, i)
		}
		result = append(result, v)
	}

	return result, nil
}

func normalizeLogicalMapValues(tableName string, count uint64, values []storage.SingleVar) ([]storage.SingleVar, error) {
	if count == 0 {
		return []storage.SingleVar{}, nil
	}

	if uint64(len(values)) < count {
		return nil, fmt.Errorf("map table %q has count=%d but only %d direct entries", tableName, count, len(values))
	}

	// Preserve the DB iteration order already captured during the raw scan.
	// We only trim stale tail entries beyond the logical Count; we do not want
	// to introduce avoidable churn in otherwise stable map exports.
	return append([]storage.SingleVar(nil), values[:count]...), nil
}

func normalizeContractVariablesLogicalState() {
	for contractName, contractItem := range contractVariables {
		for tableName, varMap := range contractItem.MapsAndLists {
			fullName := contractName + "." + tableName
			var values []storage.SingleVar
			var err error

			if looksLikeListCollection(varMap.Values) {
				values, err = normalizeLogicalListValues(fullName, varMap.Count, varMap.Values)
			} else {
				values, err = normalizeLogicalMapValues(fullName, varMap.Count, varMap.Values)
			}

			if err != nil {
				panic(err)
			}

			varMap.Values = values
			varMap.Count = uint64(len(values))
			contractItem.MapsAndLists[tableName] = varMap
		}

		contractVariables[contractName] = contractItem
	}
}

func readLogicalCount(conn *rocksdb.Connection, prefix []byte) uint64 {
	count, err := conn.GetAsBigInt(storage.CountKey(prefix))
	if err != nil {
		panic(err)
	}
	return count.Uint64()
}

func collectLogicalPrefixedEntries(conn *rocksdb.Connection, prefix []byte, count uint64) []rawKv {
	if count == 0 {
		return []rawKv{}
	}

	countKey := storage.CountKey(prefix)
	result := make([]rawKv, 0, count)
	conn.VisitPrefix(prefix, func(key []byte, value []byte) bool {
		if bytes.Equal(key, countKey) {
			return true
		}

		result = append(result, rawKv{key: key, value: value})
		return uint64(len(result)) < count
	})

	if uint64(len(result)) != count {
		panic(fmt.Sprintf("prefix %q expected %d live entries, got %d", string(prefix), count, len(result)))
	}

	return result
}
