package main

import (
	"bytes"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
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
			if bytes.HasPrefix(key, secondaryKey) {
				secondaryKey = k
				tokenSymbol = t
			}
		}

		if secondaryKey == nil {
			panic("Token is unknown")
		}

		secondaryKey = bytes.TrimPrefix(key, secondaryKey)
		secondaryKey = bytes.Join([][]byte{{34}, secondaryKey}, []byte{})
		address := io.Deserialize[*cryptography.Address](secondaryKey, &cryptography.Address{})

		br := *io.NewBinReaderFromBuf(value)
		number := br.ReadNumber()

		return Balances.String() + ": " + tokenSymbol + ": " + address.String() + " = " + number.String(), true
	}

	return string(key) + ": " + string(value), false
}
