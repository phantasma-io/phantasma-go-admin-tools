package storage

import (
	"bytes"
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/domain/types"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

type KeyValueReader struct {
	originalKeyOrValue []byte
	keyOrValue         []byte
}

func KeyValueReaderNew(keyOrValue []byte) *KeyValueReader {

	return &KeyValueReader{originalKeyOrValue: keyOrValue, keyOrValue: keyOrValue}
}

func (k *KeyValueReader) GetRemainder() []byte {
	return k.keyOrValue
}

func (k *KeyValueReader) TrimPrefix(p []byte) {
	k.keyOrValue = bytes.TrimPrefix(k.keyOrValue, []byte(p))
}

func (k *KeyValueReader) ReadBytes(hasLengthPrefix bool) []byte {
	var b []byte

	if hasLengthPrefix {
		br := *io.NewBinReaderFromBuf(k.keyOrValue)
		b = br.ReadVarBytes()
		k.keyOrValue = k.keyOrValue[br.Count:]
	} else {
		b = k.keyOrValue
		k.keyOrValue = nil
	}

	return b
}

func (k *KeyValueReader) ReadAddress(hasLengthPrefix bool) *cryptography.Address {
	var prefixedAddress []byte
	if hasLengthPrefix {
		prefixedAddress = k.keyOrValue
		k.keyOrValue = k.keyOrValue[cryptography.Length+1:]
	} else {
		prefixedAddress = bytes.Join([][]byte{{34}, k.keyOrValue}, []byte{})
		k.keyOrValue = k.keyOrValue[cryptography.Length:]
	}

	a := io.Deserialize[*cryptography.Address](prefixedAddress)

	return a
}

func (k *KeyValueReader) ReadString(hasLengthPrefix bool) string {
	var s string

	if hasLengthPrefix {
		br := *io.NewBinReaderFromBuf(k.keyOrValue)
		s = br.ReadString()
		k.keyOrValue = k.keyOrValue[br.Count:]
	} else {
		s = string(k.keyOrValue)
		k.keyOrValue = nil
	}

	return s
}

func (k *KeyValueReader) ReadOneOfStrings(options [][]byte, sep []byte) string {

	for _, s := range options {

		withSep := s
		if len(sep) > 0 {
			withSep = append(sep, withSep...)
		}

		if bytes.HasPrefix(k.keyOrValue, withSep) {
			k.keyOrValue = k.keyOrValue[len(withSep):]
			return string(s)
		}
	}

	return ""
}

func (k *KeyValueReader) ReadBigInt(hasLengthPrefix bool) *big.Int {
	var n *big.Int

	if hasLengthPrefix {
		br := *io.NewBinReaderFromBuf(k.keyOrValue)
		n = util.BigIntFromCsharpOrPhantasmaByteArray(br.ReadVarBytes())
		k.keyOrValue = k.keyOrValue[br.Count:]
	} else {
		n = util.BigIntFromCsharpOrPhantasmaByteArray(k.keyOrValue)
		k.keyOrValue = nil
	}

	return n
}

func (k *KeyValueReader) ReadTimestamp() *types.Timestamp {
	var t *types.Timestamp
	br := *io.NewBinReaderFromBuf(k.keyOrValue)
	t = br.ReadTimestamp()
	k.keyOrValue = nil
	return t
}
