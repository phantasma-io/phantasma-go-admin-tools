package main

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/domain/stake"
	phaio "github.com/phantasma-io/phantasma-go/pkg/io"
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
	} else if appOpts.DumpTokenSymbols {
		vr := storage.KeyValueReaderNew(value)
		return storage.KeyValue{Key: "Symbol", Value: vr.ReadString(true)}, true
	} else if appOpts.DumpStakingClaims {
		energyClaim := phaio.Deserialize[*stake.EnergyClaim_S](value)
		return storage.KeyValueJson{Key: keyAlt, Value: energyClaim}, true
	} else if appOpts.DumpStakes {
		energyStake := phaio.Deserialize[*stake.EnergyStake_S](value)
		return storage.KeyValueJson{Key: keyAlt, Value: energyStake}, true
	} else if appOpts.DumpStakingLeftovers {
		vr := storage.KeyValueReaderNew(value)
		return storage.KeyValueJson{Key: keyAlt, Value: vr.ReadBigInt(true).String()}, true
	} else if appOpts.DumpStakingMasterAge || appOpts.DumpStakingMasterClaims {
		vr := storage.KeyValueReaderNew(value)
		return storage.KeyValueJson{Key: keyAlt, Value: vr.ReadTimestamp()}, true
	} else if appOpts.DumpTransactions {
		kr := storage.KeyValueReaderNew(key)
		kr.TrimPrefix(Txs.Bytes())
		txHash := kr.GetRemainder()

		if len(txHash) < 33 { // It's some garbage in db
			return storage.KeyValue{}, false
		}

		blockHash, err := connection.Get(append(TxBlockMap.Bytes(), txHash...))
		if err != nil {
			panic(err)
		}

		blockHash = blockHash[1:] // First byte is length
		slices.Reverse(blockHash) // Hash is stored in reversed order.

		// Looking for block height by its hash
		ok, height := FindBlockNumberByHash(blockHash)
		if !ok {
			panic("Cannot get block height: " + height)
		}

		txHash = txHash[1:]    // First byte is length
		slices.Reverse(txHash) // Hash is stored in reversed order.

		vr := storage.KeyValueReaderNew(value)
		tx := vr.GetRemainder()

		if appOpts.Decompress {
			flateReader := flate.NewReader(bytes.NewReader(tx))
			txDecompressed, err := io.ReadAll(flateReader)
			if err != nil {
				panic(err)
			}
			tx = txDecompressed
		}
		heightInt, _ := strconv.ParseInt(height, 10, 64)
		return storage.Tx{
			TxHash:          strings.ToUpper(hex.EncodeToString(txHash)), // ToUpper() to make things easier with current explorer
			TxHashB64:       base64.StdEncoding.EncodeToString(txHash),
			BlockHashB64:    base64.StdEncoding.EncodeToString(blockHash),
			BlockHeight:     height,
			BlockHeightUint: uint64(heightInt),
			TxBytesB64:      base64.StdEncoding.EncodeToString(tx)}, true
	} else if appOpts.DumpBalances {
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
		amount := vr.ReadBigInt(false).String()

		return storage.BalanceFungible{TokenSymbol: string(tokenSymbol),
			Address: address.String(),
			Amount:  amount}, true
	} else if appOpts.DumpBalancesNft {
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
			return storage.BlockHeightAndHash{
				Height:  keyAlt,
				Hash:    strings.ToUpper(hex.EncodeToString(value)), // ToUpper() to make things easier with current explorer
				HashB64: base64.StdEncoding.EncodeToString(value)}, true
		} else if appOpts.DumpBlocks {
			block, err := connection.Get(append(Blocks.Bytes(), value...))
			if err != nil {
				panic(err)
			}

			flateReader := flate.NewReader(bytes.NewReader(block))
			blockDecompressed, err := io.ReadAll(flateReader)
			if err != nil {
				panic(err)
			}

			blockReader := storage.KeyValueReaderNew(blockDecompressed)
			blockReader.ReadBigInt(true)

			br := *phaio.NewBinReaderFromBuf(blockDecompressed)
			br.ReadVarBytes()
			timestamp := br.ReadTimestamp().Value

			if appOpts.Decompress {
				block = blockDecompressed
			}

			a := big.NewInt(0)
			a.SetString(keyAlt, 10)
			ok, blockHash := FindBlockHashByNumber(a)
			if !ok {
				panic("cannot find block hash for height " + keyAlt)
			}

			return storage.Block{Height: keyAlt,
				Hash:      base64.StdEncoding.EncodeToString(blockHash),
				Timestamp: timestamp,
				Bytes:     base64.StdEncoding.EncodeToString(block)}, true
		}
	}

	return storage.KeyValue{Key: string(key), Value: string(value)}, false
}
