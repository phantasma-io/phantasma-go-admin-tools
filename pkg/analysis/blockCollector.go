package analysis

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/util"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var one = big.NewInt(1)

func getBlocksInBatch(startHeight, groupSize *big.Int, clients []rpc.PhantasmaRPC) []response.BlockResult {
	var wg sync.WaitGroup
	res := make([]response.BlockResult, groupSize.Int64())

	endHeight := new(big.Int).Add(startHeight, groupSize)

	i := 0
	for h := new(big.Int).Set(startHeight); h.Cmp(endHeight) < 0; h.Add(h, one) {
		wg.Add(1)

		var capturedH big.Int
		capturedH.Set(h)
		go func(heightToFetch string, index int) {
			defer wg.Done()

			client := clients[rand.Intn(len(clients))]

			r, err := util.RetryHelper(func() (any, error) {
				return client.GetBlockByHeight("main", heightToFetch)
			}, 10, time.Duration(100*float64(time.Millisecond)), time.Duration(1000*float64(time.Millisecond)))

			if err != nil {
				panic("GetBlockByHeight call failed! Error: " + err.Error())
			}

			res[index] = r.(response.BlockResult)
		}(h.String(), i)

		i += 1
	}

	wg.Wait()

	return res
}

func findLatestLoadedBlock(path string) *big.Int {
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

	jsonString, _ := json.Marshal(block)
	os.WriteFile(blockPath, jsonString, 0644)
}

func GetAllBlocks(outputFolder string, clients []rpc.PhantasmaRPC) []string {

	latestLoaded := findLatestLoadedBlock(outputFolder)
	continueFrom := new(big.Int).Add(latestLoaded, one)
	fmt.Printf("Continue from block %s", continueFrom.String())

	addresses := []string{}

	chainHeight, err := clients[0].GetBlockHeight("main")
	if err != nil {
		panic("GetBlockHeight call failed! Error: " + err.Error())
	}

	groupSize := big.NewInt(30)
	blocksNotReported := 0

	start := time.Now()

	for h := continueFrom; h.Cmp(chainHeight) <= 0; h.Add(h, groupSize) {
		if blocksNotReported >= reportEveryNBlocks {
			elapsed := time.Since(start)
			fmt.Printf("Processed %s blocks in %f seconds, %f blocks per second\n", h, elapsed.Seconds(), float64(blocksNotReported)/elapsed.Seconds())
			blocksNotReported = 0
			start = time.Now()
		}

		blocks := getBlocksInBatch(h, groupSize, clients)

		for _, b := range blocks {
			storeBlock(outputFolder, b)
		}

		blocksNotReported += len(blocks)
	}

	elapsed := time.Since(start)
	fmt.Printf("Stored %s blocks in %f seconds\n", chainHeight, elapsed.Seconds())

	return addresses
}
