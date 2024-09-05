package main

import (
	"bytes"
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
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

func DumpRow(key []byte, value []byte, subkeys [][]byte, panicOnUnknownSubkey bool) (fmt.Stringer, bool) {
	if bytes.HasPrefix(key, Balances.Bytes()) {
		tokenSymbol, keyReminder := MatchWithAnySubKey(key, Balances.Bytes(), subkeys, []byte{'.'})

		if tokenSymbol == nil {
			if !panicOnUnknownSubkey {
				return storage.KeyValue{}, false
			}
			key = bytes.TrimPrefix(key, Balances.Bytes())

			// Try to show first 4 symbols of unknown token symbol
			panic("Token is unknown: '" + string(key[0:4]) + "'")
		}

		address := storage.ReadAddressWithoutLengthByte(keyReminder)
		amount := storage.ReadBigIntWithoutLengthByte(value)

		return storage.Balance{TokenSymbol: string(tokenSymbol),
			Address: address.String(),
			Amount:  amount}, true
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
