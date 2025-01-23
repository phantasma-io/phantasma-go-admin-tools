package main

import (
	"bytes"
	"encoding/hex"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

func DumpRow(connection *rocksdb.Connection, key []byte, keyAlt string, value []byte, subkeys1 [][]byte, addresses []string, panicOnUnknownSubkey bool) (storage.Exportable, bool) {
	if appOpts.DumpAddresses {
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
	} else if appOpts.DumpBlockHashes || appOpts.DumpBlocks {
		if appOpts.DumpBlockHashes {
			value = value[1:]     // First byte is length
			slices.Reverse(value) // Hash is stored in reversed order.
			return storage.KeyValue{Key: keyAlt, Value: hex.EncodeToString(value)}, true
		} else if appOpts.DumpBlocks {
			block, err := connection.Get(append(Blocks.Bytes(), value...))
			if err != nil {
				panic(err)
			}
			return storage.KeyValue{Key: keyAlt, Value: hex.EncodeToString(block)}, true
		}
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
