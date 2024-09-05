package main

import (
	"bytes"
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

func CombineKeys(baseKey, subKey, sep []byte) []byte {
	if len(sep) == 0 {
		return append(baseKey, subKey...)
	} else {
		key := append(baseKey, sep...)
		key = append(key, subKey...)
		return key
	}
}

func MatchWithAnySubKey(key, baseKey []byte, subkeys [][]byte, sep []byte) ([]byte, []byte) {
	for _, sk := range subkeys {
		combinedKey := CombineKeys(baseKey, sk, sep)
		if bytes.HasPrefix(key, combinedKey) {
			if len(combinedKey) == len(key) {
				return sk, nil
			} else {
				return sk, bytes.TrimPrefix(key, combinedKey)
			}
		}
	}

	return nil, nil
}

func MatchWithAnyAddressKey(key []byte, keys [][]byte) bool {
	for _, k := range keys {
		a, _ := cryptography.FromString(string(k))
		if bytes.Compare(key, a.Bytes()) == 0 {
			return true
		}
	}

	return false
}

func DumpRow(key []byte, value []byte, subkeys1, subkeys2 [][]byte, panicOnUnknownSubkey bool) (fmt.Stringer, bool) {
	if bytes.HasPrefix(key, Balances.Bytes()) {
		tokenSymbol, keyReminder := MatchWithAnySubKey(key, Balances.Bytes(), subkeys1, []byte{'.'})

		if tokenSymbol == nil {
			if !panicOnUnknownSubkey {
				return storage.KeyValue{}, false
			}
			key = bytes.TrimPrefix(key, Balances.Bytes())

			// Try to show first 4 symbols of unknown token symbol
			panic("Token is unknown: '" + string(key[0:4]) + "'")
		}

		if len(subkeys2) > 0 && !MatchWithAnyAddressKey(keyReminder, subkeys2) {
			return storage.KeyValue{}, false
		}

		address := storage.ReadAddressWithoutLengthByte(keyReminder)
		amount := storage.ReadBigIntWithoutLengthByte(value)

		return storage.Balance{TokenSymbol: string(tokenSymbol),
			Address: address.String(),
			Amount:  amount}, true
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
