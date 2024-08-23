package format

import "fmt"

func PrintBytesInChunks(b []byte, chunkSize int) {
	for i := 0; i < len(b); i += chunkSize {
		end := i + chunkSize

		if end > len(b) {
			end = len(b)
		}

		fmt.Println(b[i:end])
	}
}
