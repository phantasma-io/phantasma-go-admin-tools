package analysis

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/disasm"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

func ScriptsAnalyzer(script string, offset, protocol uint, debugLogging bool, calls map[string]uint, clients []rpc.PhantasmaRPC) {
	bytes, err := hex.DecodeString(script)
	if err != nil {
		panic(err)
	}

	currentOffset := disasm.ExtractMethodCallsEx(bytes[offset:], protocol, debugLogging, calls, clients)
	if debugLogging {
		fmt.Println("currentOffset: ", currentOffset)
	}
}

func TxScriptsAnalyzer(blocksFolder string, debugLogging bool, clients []rpc.PhantasmaRPC) {
	continueFrom := big.NewInt(1)
	chainHeight := FindLatestCachedBlock(blocksFolder)
	fmt.Printf("Continue from block %s, latest available block: %s\n", continueFrom.String(), chainHeight.String())

	calls := make(map[string]uint)

	blocksNotReported := 0

	start := time.Now()
	startIteration := time.Now()

	latestProcessed := big.NewInt(0)
	for h := continueFrom; h.Cmp(chainHeight) <= 0; h.Add(h, one) {
		latestProcessed = h
		const reportEveryNBlocks = 100000
		if blocksNotReported >= reportEveryNBlocks {
			elapsed := time.Since(startIteration)
			fmt.Printf("Processed %s blocks in %f seconds, %f blocks per second\n", h, elapsed.Seconds(), float64(blocksNotReported)/elapsed.Seconds())
			blocksNotReported = 0
			startIteration = time.Now()
		}

		b := readBlock(blocksFolder, h.String())

		for _, t := range b.Txs {
			if t.StateIsFault() {
				continue
			}

			if debugLogging {
				fmt.Printf("Processing tx %s [%d v%d]\n", t.Hash, t.BlockHeight, b.Protocol)
			}

			s := t.Script

			ScriptsAnalyzer(s, 0, b.Protocol, debugLogging, calls, clients)
		}

		blocksNotReported += 1
	}

	fmt.Printf("Processed %s blocks in %f seconds\n", latestProcessed, time.Since(start).Seconds())

	for k, n := range calls {
		fmt.Printf("%s: %d\n", k, n)
	}
}
