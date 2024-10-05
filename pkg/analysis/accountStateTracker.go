package analysis

import (
	"encoding/hex"
	"fmt"
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

type StakeClaimType uint

const (
	Normal      StakeClaimType = 1
	MarketEvent StakeClaimType = 2
	SmReward    StakeClaimType = 3
)

type AccountState struct {
	Tx    response.TransactionResult
	State response.AccountResult
}

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

func applyEventToAccountState(e response.EventResult,
	previousEvent *response.EventResult,
	a *response.AccountResult,
	processingDirection ProcessingDirection, tx string, stakeClaimType StakeClaimType) {

	eventKind := event.Unknown
	eventKind.SetString(e.Kind)

	var eventData *event.TokenEventData
	if !eventKind.IsTokenEvent() {
		return
	}

	// Decode event data into event.TokenEventData structure
	decoded, _ := hex.DecodeString(e.Data)
	eventData = io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

	if e.Address != a.Address {
		return
	}

	t := GetChainToken(eventData.Symbol)

	tokenBalance := a.GetTokenBalance(t)

	currentSoulStaked := big.NewInt(0)
	if a.Stake != "" {
		currentSoulStaked.SetString(a.Stake, 10)
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
				if previousEvent == nil || previousEvent.Data != e.Data { // Checking for duplicated stake event (workaround for chain bug)
					tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()
					if stakeClaimType == Normal { // We need to exclude staking related to events like "OrderFilled" or "OrderBid"
						a.Stake = amountAdd(currentSoulStaked, eventData.Value, processingDirection).String()
					}

				} else {
					fmt.Println("Check tx: " + tx)
				}
			} else {
				// For KCAL we stake amount which equals to max fee value
				tokenBalance.Amount = amountSub(currentAmount, eventData.Value, processingDirection).String()
			}

		case event.TokenClaim:
			if t.IsStakable() { // We assume it's SOUL token
				tokenBalance.Amount = amountAdd(currentAmount, eventData.Value, processingDirection).String()
				if stakeClaimType == Normal { // We need to exclude events related to SM rewards claiming and also claiming related to market events (payment when author's nft is being sold generates claim event)
					a.Stake = amountSub(currentSoulStaked, eventData.Value, processingDirection).String()
				}
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

func applyEventsToAccountState(es []response.EventResult, a *response.AccountResult, processingDirection ProcessingDirection, tx string) {
	stakeClaimType := Normal
	for _, e := range es {
		if e.Address != a.Address {
			continue
		}

		eventKind := event.Unknown
		eventKind.SetString(e.Kind)

		if eventKind.IsMarketEvent() {
			// Decode event data into event.MarketEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData := io.Deserialize[*event.MarketEventData](decoded, &event.MarketEventData{})

			if eventData.QuoteSymbol == "SOUL" {
				stakeClaimType = MarketEvent
				break
			}
		} else if eventKind == event.TokenMint {
			// Decode event data into event.TokenEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData := io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

			if eventData.Symbol == "SOUL" {
				// TODO needs better detection but for now should work
				stakeClaimType = SmReward
				break
			}
		}
	}

	for ei, e := range es {
		var previousEvent *response.EventResult
		if ei > 0 {
			previousEvent = &es[ei-1]
			if previousEvent.Address != a.Address {
				previousEvent = nil
			}
		}

		applyEventToAccountState(e,
			previousEvent,
			a,
			processingDirection, tx, stakeClaimType)
	}
}

func applyTransactionToAccountState(tx response.TransactionResult, a *response.AccountResult, processingDirection ProcessingDirection) {
	// Skip failed trasactions
	if !tx.StateIsSuccess() {
		// State din't change, saving previous one
		return
	}

	applyEventsToAccountState(tx.Events, a, processingDirection, tx.Hash)
}

// TODO Work in progress
// txs should be ordered from first tx to last
// processingDirection == Forward: We are moving from first tx to last
// processingDirection == Backward: We are moving from last tx to first
// TrackAccountStateByEvents: Modifies account to the latest state using events in transactions, also returns account state array for each transaction
func TrackAccountStateByEvents(txs []response.TransactionResult,
	account *response.AccountResult, processingDirection ProcessingDirection) []AccountState {

	perTxAccountBalances := make([]AccountState, len(txs), len(txs))

	for i := range txs {
		var tx response.TransactionResult
		if processingDirection == Forward {
			tx = txs[i]
		} else {
			tx = txs[len(txs)-1-i]
		}

		applyTransactionToAccountState(tx, account, processingDirection)
		a := account.Clone()
		perTxAccountBalances[i].Tx = tx
		perTxAccountBalances[i].State = *a
	}

	return perTxAccountBalances
}

func TrackAccountStateByEventsAndCurrentState(txs []response.TransactionResult,
	account *response.AccountResult, processingDirection ProcessingDirection) []AccountState {

	perTxAccountBalances := make([]AccountState, len(txs)+1, len(txs)+1)

	for i := range txs {
		var tx response.TransactionResult
		if processingDirection == Forward {
			tx = txs[i]
		} else {
			tx = txs[len(txs)-1-i]
		}

		a := account.Clone()
		if processingDirection == Forward {
			perTxAccountBalances[i].Tx = tx
			perTxAccountBalances[i].State = *a
		} else {
			perTxAccountBalances[len(txs)-i].Tx = tx
			perTxAccountBalances[len(txs)-i].State = *a
		}

		applyTransactionToAccountState(tx, account, processingDirection)
	}

	// We are not setting tx here because it's an initial state of account without associated tx
	a := account.Clone()
	if processingDirection == Forward {
		perTxAccountBalances[len(txs)].State = *a
	} else {
		perTxAccountBalances[0].State = *a
	}

	return perTxAccountBalances
}
