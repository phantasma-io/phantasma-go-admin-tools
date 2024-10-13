package analysis

import (
	"fmt"
	"math/big"
	"time"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

const reportEveryNBlocks = 1000

func GetAllKnownAddresses(client rpc.PhantasmaRPC) []string {
	addresses := []string{}

	chainHeight, err := client.GetBlockHeight("main")
	if err != nil {
		panic("GetBlockHeight call failed! Error: " + err.Error())
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

		blocks := getBlocksInBatch(h, groupSize, client)

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
