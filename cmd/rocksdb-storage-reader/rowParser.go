package main

import (
	"bytes"
	"encoding/json"

	"github.com/phantasma-io/phantasma-go/pkg/domain/contract"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

func ParseRow(key []byte, value []byte) (string, bool) {
	if bytes.HasPrefix(key, []byte("GHOST.serie")) {
		series := io.Deserialize[*contract.TokenSeries](value)

		// Test serialization/deserialization
		// util.SerializeDeserializePrintAndCompare(&series.ABI)
		// util.SerializPrintAndCompare(series, value)

		j, err := json.Marshal(series)
		if err != nil {
			panic(err)
		}

		return string(key) + ": " + string(j), false
	}

	return string(key) + ": " + string(value), false
}
