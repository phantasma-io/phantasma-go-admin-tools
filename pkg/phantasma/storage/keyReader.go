package storage

import (
	"bytes"
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

type KeyReader struct {
	originalKey []byte
	key         []byte
}

func KeyReaderNew(key []byte) *KeyReader {

	return &KeyReader{originalKey: key, key: key}
}

func (k *KeyReader) GetKeyRemainder() []byte {
	return k.key
}

func (k *KeyReader) TrimPrefix(p []byte) {
	k.key = bytes.TrimPrefix(k.key, []byte(p))
}

func (k *KeyReader) ReadAddress(hasLengthPrefix bool) *cryptography.Address {
	var prefixedAddress []byte
	if hasLengthPrefix {
		prefixedAddress = k.key
		k.key = k.key[cryptography.Length+1:]
	} else {
		prefixedAddress = bytes.Join([][]byte{{34}, k.key}, []byte{})
		k.key = k.key[cryptography.Length:]
	}

	a := io.Deserialize[*cryptography.Address](prefixedAddress, &cryptography.Address{})

	return a
}

func (k *KeyReader) ReadString(hasLengthPrefix bool) string {
	var s string

	if hasLengthPrefix {
		br := *io.NewBinReaderFromBuf(k.key)
		s = br.ReadString()
		k.key = k.key[br.Count:]
	} else {
		s = string(k.key)
		k.key = nil
	}

	return s
}

func (k *KeyReader) ReadOneOfStrings(options [][]byte, sep []byte) string {

	for _, s := range options {

		withSep := s
		if len(sep) > 0 {
			withSep = append(sep, withSep...)
		}

		if bytes.HasPrefix(k.key, withSep) {
			k.key = k.key[len(withSep):]
			return string(s)
		}
	}

	return ""
}

func (k *KeyReader) ReadBigInt(hasLengthPrefix bool) *big.Int {
	var n *big.Int

	if hasLengthPrefix {
		br := *io.NewBinReaderFromBuf(k.key)
		n = util.BigIntFromCsharpOrPhantasmaByteArray(br.ReadVarBytes())
		k.key = k.key[br.Count:]
	} else {
		n = util.BigIntFromCsharpOrPhantasmaByteArray(k.key)
		k.key = nil
	}

	return n
}
