package analysis

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func GetAllKnownAddresses(clients []rpc.PhantasmaRPC, blockCachePath string, verbose bool) []string {
	addresses := []string{}

	var chainHeight *big.Int
	var err error

	if blockCachePath != "" {
		chainHeight = FindLatestCachedBlock(blockCachePath)

		if verbose {
			fmt.Printf("Current cache height: %s\n", chainHeight.String())
		}
	} else {
		chainHeight, err = clients[0].GetBlockHeight("main")
		if err != nil {
			panic("GetBlockHeight call failed! Error: " + err.Error())
		}

		if verbose {
			fmt.Printf("Current chain height: %s\n", chainHeight.String())
		}
	}

	groupSize := big.NewInt(10)
	blocksNotReported := 0

	start := time.Now()
	startIteration := time.Now()

	for h := big.NewInt(1); h.Cmp(chainHeight) <= 0; h.Add(h, groupSize) {
		if verbose {
			const reportEveryNBlocks = 100000
			if blocksNotReported >= reportEveryNBlocks {
				elapsed := time.Since(startIteration)
				fmt.Printf("Processed %d blocks [%s] in %f seconds, %f blocks per second\n", blocksNotReported, h, elapsed.Seconds(), float64(blocksNotReported)/elapsed.Seconds())
				blocksNotReported = 0
				startIteration = time.Now()
			}
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
					if !slices.Contains(addresses, e.Address) {
						addresses = append(addresses, e.Address)
					}
				}
			}
		}

		blocksNotReported += len(blocks)
	}

	if verbose {
		fmt.Printf("Processed %s blocks in %f minutes, collected %d addresses\n", chainHeight, time.Since(start).Minutes(), len(addresses))
	}

	return addresses
}
