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

func ReadAddressWithoutLengthByte(value []byte) *cryptography.Address {
	value = bytes.Join([][]byte{{34}, value}, []byte{})
	return io.Deserialize[*cryptography.Address](value, &cryptography.Address{})
}

func ReadBigIntWithoutLengthByte(value []byte) *big.Int {
	return util.BigIntFromCsharpOrPhantasmaByteArray(value)
}
