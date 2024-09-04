package main

import (
	"bytes"
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func GetBalanceTokenKey(tokenSubkey []byte) []byte {
	key := Balances.Bytes()
	key = append(key, '.')
	key = append(key, tokenSubkey...)

	return key
}

func DumpRow(key []byte, value []byte, subkeys [][]byte, panicOnUnknownSubkey bool) (fmt.Stringer, bool) {
	if bytes.HasPrefix(key, Balances.Bytes()) {

		var secondaryKey []byte
		var tokenSymbol []byte
		for _, t := range subkeys {
			k := GetBalanceTokenKey(t)
			if bytes.HasPrefix(key, k) {
				secondaryKey = k
				tokenSymbol = t
			}
		}

		if secondaryKey == nil {
			if !panicOnUnknownSubkey {
				return storage.KeyValue{}, false
			}
			secondaryKey = bytes.TrimPrefix(key, Balances.Bytes())

			// Try to show first 4 symbols of unknown token symbol
			panic("Token is unknown: '" + string(secondaryKey[0:4]) + "'")
		}

		secondaryKey = bytes.TrimPrefix(key, secondaryKey)
		secondaryKey = bytes.Join([][]byte{{34}, secondaryKey}, []byte{})
		address := io.Deserialize[*cryptography.Address](secondaryKey, &cryptography.Address{})

		amount := util.BigIntFromCsharpOrPhantasmaByteArray(value)

		return storage.Balance{TokenSymbol: string(tokenSymbol),
			Address: address.String(),
			Amount:  amount}, true
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
