package analysis

import (
	"encoding/hex"
	"slices"

	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func CheckIfTransactionHasEvent(tx response.TransactionResult,
	address, tokenSymbol, payloadFragment string, eventKinds []event.EventKind, showFailedTxes bool) bool {

	// Skip failed trasactions
	if !tx.StateIsSuccess() && !showFailedTxes {
		return false
	}

	for _, e := range tx.Events {
		if e.Address != address {
			continue
		}

		eventKind := event.Unknown
		eventKind.SetString(e.Kind)

		if len(eventKinds) != 0 && !slices.Contains(eventKinds, eventKind) {
			continue
		}

		var eventData *event.TokenEventData
		if eventKind.IsTokenEvent() {
			// Decode event data into event.TokenEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData = io.Deserialize[*event.TokenEventData](decoded)

			if tokenSymbol != "" && tokenSymbol != eventData.Symbol {
				continue
			}

			payloadBytes, _ := hex.DecodeString(tx.Payload)
			payload := string(payloadBytes)

			if payloadFragment != "" && payloadFragment != payload {
				continue
			}

			return true
		}
	}

	return false
}

func GetTransactionsByKind(txs []response.TransactionResult,
	address, tokenSymbol, payloadFragment string, eventKinds []event.EventKind, showFailedTxes bool) []response.TransactionResult {
	var result []response.TransactionResult

	for i := 0; i < len(txs); i++ {
		if CheckIfTransactionHasEvent(txs[i], address, tokenSymbol, payloadFragment, eventKinds, showFailedTxes) {
			result = append(result, txs[i])
		}
	}
	return result
}
