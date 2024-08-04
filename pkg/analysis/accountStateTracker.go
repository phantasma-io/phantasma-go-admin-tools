package analysis

import (
	"encoding/hex"
	"math/big"
	"slices"

	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/io"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

type ProcessingDirection uint

const (
	Forward  ProcessingDirection = 1
	Backward ProcessingDirection = 2
)

func arrayHasEmpoty(a []string) bool {
	for _, s := range a {
		if s == "" {
			return true
		}
	}
	return false
}

// amountAdd Adds value to balance's amount
func amountAdd(currentAmount *big.Int, deltaAmount *big.Int, processingDirection ProcessingDirection) *big.Int {
	if processingDirection == Forward {
		return new(big.Int).Add(currentAmount, deltaAmount)
	} else {
		return new(big.Int).Sub(currentAmount, deltaAmount)
	}
}

// amountSub Subtracts value from balance's amount
func amountSub(currentAmount *big.Int, deltaAmount *big.Int, processingDirection ProcessingDirection) *big.Int {
	if processingDirection == Forward {
		return new(big.Int).Sub(currentAmount, deltaAmount)
	} else {
		return new(big.Int).Add(currentAmount, deltaAmount)
	}
}

// idAdd Adds NFT ID to balance's IDs array
func idAdd(ids *[]string, id string, processingDirection ProcessingDirection) {
	if processingDirection == Forward {
		*ids = append(*ids, id)
	} else {
		for i, foundId := range *ids {
			if foundId == id {
				*ids = slices.Delete(*ids, i, i+1)
			}
		}
	}
}

// idRemove Removes NFT ID from balance's IDs array
func idRemove(ids *[]string, id string, processingDirection ProcessingDirection) {
	if processingDirection == Forward {
		for i, foundId := range *ids {
			if foundId == id {
				*ids = slices.Delete(*ids, i, i+1)
			}
		}
	} else {
		*ids = append(*ids, id)
	}
}

// TODO Work in progress
// txs should be ordered from first tx to last
// processingDirection == Forward: We are moving from first tx to last
// processingDirection == Backward: We are moving from last tx to first
// TrackAccountStateByEvents: Modifies account to the latest state using events in transactions, also returns account state array for each transaction
func TrackAccountStateByEvents(txs []response.TransactionResult, accountAddress string,
	account *response.AccountResult, processingDirection ProcessingDirection) *[]response.AccountResult {

	perTxAccountBalances := make([]response.AccountResult, len(txs), len(txs))

	account.Address = accountAddress

	// Calculate balance
	i := 0
	if processingDirection == Forward {
		i = 0
	} else {
		i = len(txs) - 1
	}

	for {
		if processingDirection == Forward && i == len(txs) {
			break
		} else if i < 0 {
			break
		}

		// Skip failed trasactions
		if !txs[i].StateIsSuccess() {
			// State din't change, saving previous one
			perTxAccountBalances[i] = *account

			if processingDirection == Forward {
				i += 1
			} else {
				i -= 1
			}

			continue
		}

		for ei, e := range txs[i].Events {
			eventKind := event.Unknown
			eventKind.SetString(e.Kind)

			var eventData *event.TokenEventData
			if eventKind.IsTokenEvent() {
				// Decode event data into event.TokenEventData structure
				decoded, _ := hex.DecodeString(e.Data)
				eventData = io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

				if e.Address != accountAddress {
					continue
				}

				t := GetChainToken(eventData.Symbol)

				tokenBalance := account.GetTokenBalance(t)

				currentSoulStaked := big.NewInt(0)
				if account.Stake != "" {
					currentSoulStaked.SetString(account.Stake, 10)
				}

				currentAmount, _ := big.NewInt(0).SetString(tokenBalance.Amount, 10)

				if t.IsFungible() {
					switch eventKind {
					// Processing unstaked balance
					case event.TokenReceive:
						tokenBalance.Amount = amountAdd(currentAmount, eventData.Value, processingDirection).String()
					case event.TokenSend:
						tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()

					// Process staking
					// For now we process stakes for SOUL only, ignoring isStakable() flag
					case event.TokenStake:
						if t.IsStakable() { // We assume it's SOUL token
							if ei > 0 && txs[i].Events[ei-1].Data != e.Data { // Checking for duplicated stake event (workaround for chain bug)
								tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()
								account.Stake = amountAdd(currentSoulStaked, eventData.Value, processingDirection).String()
							}
						} else {
							// For KCAL we stake amount which equals to max fee value
							tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()
						}

					case event.TokenClaim:
						if t.IsStakable() { // We assume it's SOUL token
							tokenBalance.Amount = amountAdd(currentAmount, eventData.Value, processingDirection).String()
							account.Stake = amountSub(currentSoulStaked, eventData.Value, processingDirection).String()
						} else { // KCAL claim
							tokenBalance.Amount = amountAdd(currentAmount, eventData.Value, processingDirection).String()
							// We can't properly track account.Unclaimed here, it needs to be calculated
						}

					case event.TokenBurn:
						tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()

					case event.TokenMint:
						tokenBalance.Amount = amountAdd(currentAmount, eventData.Value, processingDirection).String()
					}
				} else {
					switch eventKind {
					case event.TokenReceive:
						tokenBalance.Amount = amountAdd(currentAmount, big.NewInt(1), processingDirection).String()
						idAdd(&tokenBalance.Ids, eventData.Value.String(), processingDirection)

					case event.TokenSend:
						tokenBalance.Amount = amountSub(currentAmount, big.NewInt(1), processingDirection).String()
						idRemove(&tokenBalance.Ids, eventData.Value.String(), processingDirection)

					case event.TokenBurn:
						tokenBalance.Amount = amountSub(currentAmount, big.NewInt(1), processingDirection).String()
						idRemove(&tokenBalance.Ids, eventData.Value.String(), processingDirection)

					case event.TokenMint:
						tokenBalance.Amount = amountAdd(currentAmount, big.NewInt(1), processingDirection).String()
						idAdd(&tokenBalance.Ids, eventData.Value.String(), processingDirection)
					}
				}
			}
		}

		perTxAccountBalances[i] = *account.Clone()

		if processingDirection == Forward {
			i += 1
		} else {
			i -= 1
		}
	}

	return &perTxAccountBalances
}
