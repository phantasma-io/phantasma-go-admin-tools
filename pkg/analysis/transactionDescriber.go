package analysis

import (
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/format"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

type OrderDirection uint

const (
	Asc  OrderDirection = 1
	Desc OrderDirection = 2
)

// Work in progress
func DescribeTransaction(perTxAccountBalance *AccountState,
	address, tokenSymbol, payloadFragment string, describeFungible, describeNonfungible bool, eventKinds []event.EventKind) (string, []string, bool) {
	txInfo := fmt.Sprintf("Hash: %s", perTxAccountBalance.Tx.Hash)
	eventsInfo := []string{}

	stateIsSuccess := perTxAccountBalance.Tx.StateIsSuccess()

	// Skip failed trasactions
	if !stateIsSuccess {
		txInfo += " [FAILED]"
		return txInfo, eventsInfo, stateIsSuccess
	}

	for _, e := range perTxAccountBalance.Tx.Events {
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
			eventData = io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

			if tokenSymbol != "" && tokenSymbol != eventData.Symbol {
				continue
			}

			payloadBytes, _ := hex.DecodeString(perTxAccountBalance.Tx.Payload)
			payload := string(payloadBytes)

			if payloadFragment != "" && payloadFragment != payload {
				continue
			}

			if eventKind == event.TokenStake && eventData.Symbol != "SOUL" {
				continue
			}

			// Apply decimals to the token amount
			t := GetChainToken(eventData.Symbol)
			tokenAmount := util.ConvertDecimals(eventData.Value, int(t.Decimals))

			if t.IsFungible() && describeFungible {
				switch eventKind {
				case event.TokenStake:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).State.GetTokenBalance(t).ConvertDecimals()))

				case event.TokenClaim:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).State.GetTokenBalance(t).ConvertDecimals()))

				case event.TokenReceive:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).State.GetTokenBalance(t).ConvertDecimals()))

				case event.TokenSend:
					// We found TokenSend event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).State.GetTokenBalance(t).ConvertDecimals()))

				case event.TokenMint:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).State.GetTokenBalance(t).ConvertDecimals()))
				}
			} else if !t.IsFungible() && describeNonfungible {
				b := (*perTxAccountBalance).State.GetTokenBalance(t)
				switch eventKind {
				case event.TokenReceive:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenSend:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenBurn:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenMint:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))
				}
			}
		}
	}

	return txInfo, eventsInfo, stateIsSuccess
}

// Work in progress
func DescribeTransactions(includedTxes []response.TransactionResult, perTxAccountBalances []AccountState,
	address, tokenSymbol, payloadFragment string, describeFungible, describeNonfungible bool, eventKinds []event.EventKind, showFailedTxes bool) []string {

	var result []string

	for i, perTxAccountBalance := range perTxAccountBalances {
		includedTx := true
		if len(includedTxes) != 0 {
			includedTx = false
			for _, t := range includedTxes {
				if t.Hash == perTxAccountBalance.Tx.Hash {
					includedTx = true
				}
			}
		}

		if includedTx {
			txInfo, eventsInfo, stateIsSuccess := DescribeTransaction(&perTxAccountBalance, address, tokenSymbol, payloadFragment, describeFungible, describeNonfungible, eventKinds)
			var txBlock string
			if len(eventsInfo) != 0 || (showFailedTxes && !stateIsSuccess) {
				txBlock += fmt.Sprintf("#%03d %s %s", i+1, time.Unix(int64(perTxAccountBalance.Tx.Timestamp), 0).UTC().Format(time.RFC822), txInfo)
				for _, e := range eventsInfo {
					txBlock += fmt.Sprintf("\n\t %s", e)
				}

				result = append(result, txBlock)
			}
		}
	}
	return result
}
