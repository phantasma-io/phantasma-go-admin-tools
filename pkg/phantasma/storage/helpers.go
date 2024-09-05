package storage

import (
	"bytes"
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func ReadStringWithLengthByte(value []byte) string {
	br := *io.NewBinReaderFromBuf(value)
	return br.ReadString()
}

func ReadAddressWithLengthByte(value []byte) *cryptography.Address {
	return io.Deserialize[*cryptography.Address](value, &cryptography.Address{})
}

func ReadAddressWithoutLengthByte(value []byte) *cryptography.Address {
	value = bytes.Join([][]byte{{34}, value}, []byte{})
	return ReadAddressWithLengthByte(value)
}

func ReadBigIntWithoutLengthByte(value []byte) *big.Int {
	return util.BigIntFromCsharpOrPhantasmaByteArray(value)
}
