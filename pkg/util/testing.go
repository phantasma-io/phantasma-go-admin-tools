package util

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/format"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

func SerializPrintAndCompare[T io.Serializer](object T, reference []byte) {
	serialized := io.Serialize[T](object)

	if bytes.Compare(serialized, reference) != 0 {
		chunkSize := 10
		format.PrintBytesInChunks(serialized, chunkSize)
		fmt.Print("\n\n")
		format.PrintBytesInChunks(reference, chunkSize)

		fmt.Print("\n\n")
		j, err := json.Marshal(io.Deserialize[T](serialized))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))

		fmt.Print("\n\n")
		j, err = json.Marshal(io.Deserialize[T](reference))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))

		panic("SerializPrintAndCompare(): Bytes differ!")
	}
}

func SerializeDeserializePrintAndCompare[T io.Serializer](object T) {
	serialized1 := io.Serialize[T](object)

	io.Deserialize[T](serialized1)
	serialized2 := io.Serialize[T](object)

	if bytes.Compare(serialized1, serialized2) != 0 {
		chunkSize := 10
		format.PrintBytesInChunks(serialized1, chunkSize)
		fmt.Print("\n\n")
		format.PrintBytesInChunks(serialized2, chunkSize)

		fmt.Print("\n\n")
		j, err := json.Marshal(io.Deserialize[T](serialized1))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))

		fmt.Print("\n\n")
		j, err = json.Marshal(io.Deserialize[T](serialized2))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))

		panic("SerializeDeserializePrintAndCompare(): Bytes differ!")
	}
}
