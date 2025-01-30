package main

import (
	"bytes"
	"compress/flate"
	"io"
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

func BigIntFromString(n string) *big.Int {
	bi := big.NewInt(0)
	bi.SetString(n, 10)
	return bi
}

func Decompress(compressed []byte) []byte {
	flateReader := flate.NewReader(bytes.NewReader(compressed))
	bytesDecompressed, err := io.ReadAll(flateReader)
	if err != nil {
		panic(err)
	}
	return bytesDecompressed
}

func GetNftTokenKey(symbol, tokenId string) []byte {
	return []byte(symbol + "." + tokenId)
}

func GetTokenSeriesKey(symbol string, seriesID string) []byte {
	return []byte(symbol + ".serie" + seriesID)
}
