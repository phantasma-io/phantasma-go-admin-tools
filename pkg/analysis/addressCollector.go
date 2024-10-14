package analysis

import (
	"fmt"
	"math/big"
	"time"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

const reportEveryNBlocks = 1000

func GetAllKnownAddresses(clients []rpc.PhantasmaRPC, blockCachePath string) []string {
	addresses := []string{}

	var chainHeight *big.Int
	var err error

	if blockCachePath != "" {
		chainHeight = FindLatestCachedBlock(blockCachePath)

		fmt.Printf("Current cache height: %s\n", chainHeight.String())
	} else {
		chainHeight, err = clients[0].GetBlockHeight("main")
		if err != nil {
			panic("GetBlockHeight call failed! Error: " + err.Error())
		}

		fmt.Printf("Current chain height: %s\n", chainHeight.String())
	}

	groupSize := big.NewInt(10)
	blocksNotReported := 0

	start := time.Now()

	for h := big.NewInt(1); h.Cmp(chainHeight) <= 0; h.Add(h, groupSize) {
		if blocksNotReported >= reportEveryNBlocks {
			elapsed := time.Since(start)
			fmt.Printf("Processed %s blocks in %f seconds, %f blocks per second\n", h, elapsed.Seconds(), float64(blocksNotReported)/elapsed.Seconds())
			blocksNotReported = 0
			start = time.Now()
		}

		var blocks []response.BlockResult

		// Last group might be smaller, we need to calculate correct size of last group
		var currentGroupSize *big.Int
		currentGroupSize = groupSize
		delta := new(big.Int).Sub(chainHeight, h)
		if currentGroupSize.Cmp(delta) == 1 {
			currentGroupSize = delta
		}

		if blockCachePath != "" {
			blocks = GetBlocksInBatchFromCache(h, currentGroupSize, blockCachePath)
		} else {
			blocks = getBlocksInBatch(h, currentGroupSize, clients)
		}

		for _, b := range blocks {
			txs := b.Txs
			for _, tx := range txs {
				for _, e := range tx.Events {
					addresses = append(addresses, e.Address)
				}
			}
		}

		blocksNotReported += len(blocks)
	}

	elapsed := time.Since(start)
	fmt.Printf("Processed %s blocks in %f seconds\n", chainHeight, elapsed.Seconds())

	return addresses
}
