package storage

import "github.com/phantasma-io/phantasma-go/pkg/io"

func ReadStringWithLengthByte(value []byte) string {
	br := *io.NewBinReaderFromBuf(value)
	return br.ReadString()
}
