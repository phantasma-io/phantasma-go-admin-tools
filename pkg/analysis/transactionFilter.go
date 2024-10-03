package analysis

import (
	"encoding/hex"

	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func CheckIfTransactionHasEvent(tx response.TransactionResult,
	address, tokenSymbol, payloadFragment string, k event.EventKind) bool {

	// Skip failed trasactions
	if !tx.StateIsSuccess() {
		return false
	}

	for _, e := range tx.Events {
		if e.Address != address {
			continue
		}

		eventKind := event.Unknown
		eventKind.SetString(e.Kind)

		var eventData *event.TokenEventData
		if eventKind.IsTokenEvent() {
			// Decode event data into event.TokenEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData = io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

			if tokenSymbol != "" && tokenSymbol != eventData.Symbol {
				continue
			}

			payloadBytes, _ := hex.DecodeString(tx.Payload)
			payload := string(payloadBytes)

			if payloadFragment != "" && payloadFragment != payload {
				continue
			}

			if eventKind != k {
				continue
			}

			return true
		}
	}

	return false
}

func GetTransactionsByKind(txs []response.TransactionResult,
	address, tokenSymbol, payloadFragment string, eventKind event.EventKind) []response.TransactionResult {
	var result []response.TransactionResult

	for i := 0; i < len(txs); i++ {
		if CheckIfTransactionHasEvent(txs[i], address, tokenSymbol, payloadFragment, eventKind) {
			result = append(result, txs[i])
		}
	}
	return result
}
