package analysis

import (
	"encoding/json"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func GetBlocksInBatchFromCache(startHeight, groupSize *big.Int, blockCachePath string) []response.BlockResult {
	var wg sync.WaitGroup
	res := make([]response.BlockResult, groupSize.Int64())

	endHeight := new(big.Int).Add(startHeight, groupSize)

	i := 0
	for h := new(big.Int).Set(startHeight); h.Cmp(endHeight) < 0; h.Add(h, one) {
		res[i] = readBlock(blockCachePath, h.String())

		i += 1
	}

	wg.Wait()

	return res
}

func FindLatestCachedBlock(path string) *big.Int {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	latest := big.NewInt(0)
	for _, file := range files {
		if file.Type().IsRegular() {
			h, success := new(big.Int).SetString(file.Name(), 10)
			if success {
				if h.Cmp(latest) == 1 {
					latest = h
				}
			}
		}
	}

	return latest
}

func storeBlock(path string, block response.BlockResult) {
	blockPath := filepath.Join(path, strconv.FormatUint(uint64(block.Height), 10))

	jsonString, err := json.Marshal(block)
	if err != nil {
		panic("storeBlock call failed! Error: " + err.Error())
	}

	os.WriteFile(blockPath, jsonString, 0644)
}

func readBlock(path, height string) response.BlockResult {
	blockPath := filepath.Join(path, height)

	byteValue, err := os.ReadFile(blockPath)
	if err != nil {
		panic("readBlock call failed! Error: " + err.Error())
	}

	var b response.BlockResult
	json.Unmarshal(byteValue, &b)

	return b
}
