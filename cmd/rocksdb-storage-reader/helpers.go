package main

import (
	"math/big"
	"strconv"
)

func FindBlockNumberByHash(hash []byte) (bool, string) {
	height, ok := appOpts.blockHeightsMap[string(hash)]
	if ok {
		return ok, strconv.Itoa(height + 1)
	}
	return ok, ""
}

func FindBlockHashByNumber(height *big.Int) (bool, []byte) {
	hash, ok := appOpts.blockHeightsMap2[int(height.Uint64())-1]
	if ok {
		return ok, []byte(hash)
	}
	return ok, nil
}
