package main

import (
	"bytes"
	"encoding/json"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/domain/contract"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

func GetBalanceTokenAddressKey(address []byte, tokenSubkey []byte) []byte {
	key := CombineKeys(Balances.Bytes(), tokenSubkey, []byte{'.'})
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
