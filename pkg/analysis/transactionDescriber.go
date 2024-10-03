package analysis

import (
	"encoding/hex"
	"fmt"
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
func DescribeTransaction(tx response.TransactionResult, perTxAccountBalance *response.AccountResult,
	address, tokenSymbol, payloadFragment string, describeFungible, describeNonfungible bool) (string, []string) {
	txInfo := fmt.Sprintf("Hash: %s", tx.Hash)
	eventsInfo := []string{}

	// Skip failed trasactions
	if !tx.StateIsSuccess() {
		txInfo += " [FAILED]"
		return txInfo, eventsInfo
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
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).GetTokenBalance(t).ConvertDecimals()))

				case event.TokenClaim:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).GetTokenBalance(t).ConvertDecimals()))

				case event.TokenReceive:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).GetTokenBalance(t).ConvertDecimals()))

				case event.TokenSend:
					// We found TokenSend event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, (*perTxAccountBalance).GetTokenBalance(t).ConvertDecimals()))
				}
			} else if !t.IsFungible() && describeNonfungible {
				b := (*perTxAccountBalance).GetTokenBalance(t)
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

	return txInfo, eventsInfo
}

// Work in progress
func DescribeTransactions(txs []response.TransactionResult, includedTxes []response.TransactionResult, perTxAccountBalances []response.AccountResult, txIndexes []int,
	address, tokenSymbol, payloadFragment string, orderDirection OrderDirection, describeFungible, describeNonfungible bool) string {
	var result string

	i := 0
	if orderDirection == Asc {
		i = 0
	} else {
		i = len(txs) - 1
	}

	for {
		if orderDirection == Asc && i == len(txs) {
			break
		} else if i < 0 {
			break
		}

		includedTx := true
		if len(includedTxes) != 0 {
			includedTx = false
			for _, t := range includedTxes {
				if t.Hash == txs[i].Hash {
					includedTx = true
				}
			}
		}

		if includedTx {
			txInfo, eventsInfo := DescribeTransaction(txs[i], &perTxAccountBalances[i], address, tokenSymbol, payloadFragment, describeFungible, describeNonfungible)
			result += fmt.Sprintf("#%03d %s %s\n", txIndexes[i], time.Unix(int64(txs[i].Timestamp), 0).UTC().Format(time.RFC822), txInfo)
			for _, e := range eventsInfo {
				result += fmt.Sprintf("\t %s\n", e)
			}
		}

		if orderDirection == Asc {
			i += 1
		} else {
			i -= 1
		}
	}
	return result
}
