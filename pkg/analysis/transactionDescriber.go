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
func DescribeTransaction(state *AccountState,
	address, tokenSymbol, payloadFragment string, describeFungible, describeNonfungible bool, eventKinds []event.EventKind) (string, []string, bool) {
	txInfo := fmt.Sprintf("Hash: %s", state.Tx.Hash)
	eventsInfo := []string{}

	stateIsSuccess := state.Tx.StateIsSuccess()

	// Skip failed trasactions
	if !stateIsSuccess {
		txInfo += " [FAILED]"
		return txInfo, eventsInfo, stateIsSuccess
	}

	for _, e := range state.Tx.Events {
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

			payloadBytes, _ := hex.DecodeString(state.Tx.Payload)
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

			b := state.State.GetTokenBalance(t)

			if t.IsFungible() && describeFungible {
				switch eventKind {
				case event.TokenStake:
					// We found TokenStake event for given address
					if t.Symbol == "SOUL" {
						smLabel := ""
						if state.IsSm {
							smLabel = " *SM*"
						}

						stakeClaimType := " "
						if state.StakeClaimType == MarketEvent {
							stakeClaimType = "M"
						} else if state.StakeClaimType == SmReward {
							stakeClaimType = "S"
						}
						eventsInfo = append(eventsInfo, fmt.Sprintf("%-10s [%s] %-18s %-6s %-23s %s [%s]%s", eventKind, stakeClaimType, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals(), state.State.Stakes.ConvertDecimals(), smLabel))
					} else {
						eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals()))
					}

				case event.TokenClaim:
					// We found TokenClaim event for given address
					if t.Symbol == "SOUL" {
						smLabel := ""
						if state.IsSm {
							smLabel = " *SM*"
						}

						stakeClaimType := " "
						if state.StakeClaimType == MarketEvent {
							stakeClaimType = "M"
						} else if state.StakeClaimType == SmReward {
							stakeClaimType = "S"
						}
						eventsInfo = append(eventsInfo, fmt.Sprintf("%-10s [%s] %-18s %-6s %-23s %s [%s]%s", eventKind, stakeClaimType, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals(), state.State.Stakes.ConvertDecimals(), smLabel))
					} else {
						eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals()))
					}

				case event.TokenReceive:
					// We found TokenReceive event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals()))

				case event.TokenSend:
					// We found TokenSend event for given address
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals()))

				case event.TokenMint:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s", eventKind, tokenAmount, eventData.Symbol, payload, b.ConvertDecimals()))
				}
			} else if !t.IsFungible() && describeNonfungible {
				switch eventKind {
				case event.TokenReceive:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenSend:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenBurn:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenMint:
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-14s %-18s %-6s %-23s %s [%s]", eventKind, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenStake:
					stakeClaimType := " "
					if state.StakeClaimType == MarketEvent {
						stakeClaimType = "M"
					} else if state.StakeClaimType == SmReward {
						stakeClaimType = "S"
					}
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-10s [%s] %-18s %-6s %-23s %s [%s]", eventKind, stakeClaimType, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))

				case event.TokenClaim:
					stakeClaimType := " "
					if state.StakeClaimType == MarketEvent {
						stakeClaimType = "M"
					} else if state.StakeClaimType == SmReward {
						stakeClaimType = "S"
					}
					eventsInfo = append(eventsInfo, fmt.Sprintf("%-10s [%s] %-18s %-6s %-23s %s [%s]", eventKind, stakeClaimType, "1 NFT "+format.ShortenNftId(tokenAmount), eventData.Symbol, payload, b.ConvertDecimals(), format.NftIdsToString(b.Ids, ", ", true)))
				}
			}
		}
	}

	return txInfo, eventsInfo, stateIsSuccess
}

// Work in progress
func DescribeTransactions(includedTxes []response.TransactionResult, states []AccountState,
	address, tokenSymbol, payloadFragment string, describeFungible, describeNonfungible bool, eventKinds []event.EventKind, showFailedTxes bool) []string {

	var result []string

	for i, s := range states {
		includedTx := true
		if len(includedTxes) != 0 {
			includedTx = false
			for _, t := range includedTxes {
				if t.Hash == s.Tx.Hash {
					includedTx = true
				}
			}
		}

		if includedTx {
			txInfo, eventsInfo, stateIsSuccess := DescribeTransaction(&s, address, tokenSymbol, payloadFragment, describeFungible, describeNonfungible, eventKinds)
			var txBlock string
			if len(eventsInfo) != 0 || (showFailedTxes && !stateIsSuccess) {
				txBlock += fmt.Sprintf("#%03d %s %s", i+1, time.Unix(int64(s.Tx.Timestamp), 0).UTC().Format(time.RFC822), txInfo)
				for _, e := range eventsInfo {
					txBlock += fmt.Sprintf("\n\t %s", e)
				}

				result = append(result, txBlock)
			}
		}
	}
	return result
}
