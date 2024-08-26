package util

import (
	"bytes"
	"strconv"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

func ParseAsAddress(b []byte, panicOnError bool) (bool, string) {
	if len(b) != 32 && len(b) != 33 {
		if panicOnError {
			panic("Incorrect address bytes length: " + strconv.Itoa(len(b)))
		} else {
			return false, ""
		}
	}

	addressBytes := b

	if len(b) == 32 {
		addressBytes = bytes.Join([][]byte{{34}, addressBytes}, []byte{})
	}

	a := io.Deserialize[*cryptography.Address](addressBytes, &cryptography.Address{})
	return true, a.String()
}

func ParseAsHash(b []byte, panicOnError bool) (bool, string) {
	if len(b) != 33 {
		if panicOnError {
			panic("Incorrect hash bytes length: " + strconv.Itoa(len(b)))
		} else {
			return false, ""
		}
	}

	hashBytes := b
	h := io.Deserialize[*cryptography.Hash](hashBytes, &cryptography.Hash{})
	return true, h.String()
}
