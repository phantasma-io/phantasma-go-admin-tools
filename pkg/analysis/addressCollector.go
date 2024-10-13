package analysis

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func RetryHelper[T any](fn func() (T, error), maxRetry int, startBackoff, maxBackoff time.Duration) (T, error) {

	for attempt := 0; ; attempt++ {
		result, err := fn()
		if err == nil {
			return result, err
		}

		if attempt == maxRetry-1 {
			return result, err
		}

		fmt.Printf("Retrying after %s\n", startBackoff)
		time.Sleep(startBackoff)
		if maxBackoff == 0 || startBackoff < maxBackoff {
			startBackoff *= 2
		}
	}
}

const reportEveryNBlocks = 1000

var one = big.NewInt(1)

func getBlocksInBatch(startHeight, groupSize *big.Int, client rpc.PhantasmaRPC) []response.BlockResult {
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

			r, err := RetryHelper(func() (any, error) {
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
