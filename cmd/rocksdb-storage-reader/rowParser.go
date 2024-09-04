package main

import (
	"bytes"
	"encoding/json"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/domain/contract"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func GetBalanceTokenKey(tokenSubkey string) []byte {
	key := Balances.Bytes()
	key = append(key, []byte(tokenSubkey)...)

	return key
}

func GetBalanceTokenAddressKey(address []byte, tokenSubkey string) []byte {
	key := GetBalanceTokenKey(tokenSubkey)
	key = append(key, address...)

	return key
}

func GetAccountAddressMapKey(address []byte) []byte {
	key := AccountAddressMap.Bytes()
	key = append(key, address...)

	return key
}

func ParseRow(key []byte, value []byte) (string, bool) {
	if bytes.HasPrefix(key, AccountAddressMap.Bytes()) {

		secondaryKey := bytes.TrimPrefix(key, AccountAddressMap.Bytes())
		address := io.Deserialize[*cryptography.Address](secondaryKey, &cryptography.Address{})
		// TODO value has some prefix, to fix
		return AccountAddressMap.String() + "." + address.String() + ": " + string(value), true
	}

	if bytes.HasPrefix(key, Balances.Bytes()) {

		var secondaryKey []byte
		var tokenSymbol string
		for _, t := range KnowSubKeys[Balances] {
			k := GetBalanceTokenKey(t)
			if bytes.HasPrefix(key, k) {
				secondaryKey = k
				tokenSymbol = t
			}
		}

		if secondaryKey == nil {
			secondaryKey = bytes.TrimPrefix(key, Balances.Bytes())

			// Try to show first 4 symbols of unknown token symbol
			panic("Token is unknown: '" + string(secondaryKey[0:4]) + "'")
		}

		secondaryKey = bytes.TrimPrefix(key, secondaryKey)
		secondaryKey = bytes.Join([][]byte{{34}, secondaryKey}, []byte{})
		address := io.Deserialize[*cryptography.Address](secondaryKey, &cryptography.Address{})

		number := util.BigIntFromCsharpOrPhantasmaByteArray(value)

		return Balances.String() + ": " + tokenSymbol + ": " + address.String() + " = " + number.String(), true
	}

	if bytes.HasPrefix(key, []byte("GHOST.serie")) {
		series := io.Deserialize[*contract.TokenSeries](value, &contract.TokenSeries{})

		// Test serialization/deserialization
		// util.SerializeDeserializePrintAndCompare(&series.ABI)
		// util.SerializPrintAndCompare(series, value)

		j, err := json.Marshal(series)
		if err != nil {
			panic(err)
		}

		return string(key) + ": " + string(j), false
	}

	return string(key) + ": " + string(value), false
}
