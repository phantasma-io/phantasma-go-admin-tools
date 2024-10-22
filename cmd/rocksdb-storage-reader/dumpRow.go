package main

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

func DumpRow(connection *rocksdb.Connection, key []byte, value []byte, subkeys1 [][]byte, addresses []string, panicOnUnknownSubkey bool) (fmt.Stringer, bool) {
	if bytes.HasPrefix(key, AccountAddressMap.Bytes()) {
		kr := storage.KeyValueReaderNew(key)
		kr.TrimPrefix(AccountAddressMap.Bytes())

		if string(kr.GetRemainder()) == "{count}" {
			return storage.KeyValue{}, false
		}

		address := kr.ReadAddress(true)

		if len(addresses) > 0 {
			if !slices.Contains(addresses, address.String()) {
				return storage.KeyValue{}, false
			}
		}

		vr := storage.KeyValueReaderNew(value)
		name := vr.ReadString(true)

		return storage.Address{Address: address.String(), Name: name}, true
	} else if bytes.HasPrefix(key, Balances.Bytes()) {
		kr := storage.KeyValueReaderNew(key)
		kr.TrimPrefix(Balances.Bytes())

		tokenSymbol := kr.ReadOneOfStrings(subkeys1, []byte{'.'})
		if tokenSymbol == "" {
			return storage.KeyValue{}, false
		}

		address := kr.ReadAddress(false)

		if len(addresses) > 0 {
			if !slices.Contains(addresses, address.String()) {
				return storage.KeyValue{}, false
			}
		}

		vr := storage.KeyValueReaderNew(value)
		amount := vr.ReadBigInt(false)

		return storage.BalanceFungible{TokenSymbol: string(tokenSymbol),
			Address: address.String(),
			Amount:  amount}, true
	} else if bytes.HasPrefix(key, Ids.Bytes()) {
		// OwnershipSheet: '.ids.symbol' + address.ToByteArray()

		kr := storage.KeyValueReaderNew(key)
		kr.TrimPrefix(Ids.Bytes())

		tokenSymbol := kr.ReadOneOfStrings(subkeys1, []byte{'.'})
		if tokenSymbol == "" {
			return storage.KeyValue{}, false
		}

		address := kr.ReadAddress(false)

		if len(addresses) > 0 {
			if !slices.Contains(addresses, address.String()) {
				return storage.KeyValue{}, false
			}
		}

		if string(kr.GetRemainder()) == "{count}" {
			return storage.KeyValue{}, false
		}

		tokenId := kr.ReadBigInt(true)

		return storage.BalanceNonFungibleSingleRow{TokenSymbol: tokenSymbol,
			Address: address.String(),
			Id:      tokenId.String()}, true
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
