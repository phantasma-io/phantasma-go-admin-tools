package storage

import (
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/util"
)

// Some utils copied from Phantasma.Core.Storage.Context.CollectionUtils

var count_prefix []byte = []byte("{count}")
var element_begin_prefix []byte = []byte{'<'}
var element_end_prefix []byte = []byte{'>'}

func CountKey(baseKey []byte) []byte {
	return append(baseKey, count_prefix...)
}

func ElementKey(baseKey []byte, index *big.Int) []byte {
	var right []byte

	if index.BitLen() == 0 {
		right = append(element_begin_prefix, 0)
	} else {
		right = append(element_begin_prefix, util.BigIntToPhantasmaByteArray(index)...)
	}

	right = append(right, element_end_prefix...)

	return append(baseKey, right...)
}
