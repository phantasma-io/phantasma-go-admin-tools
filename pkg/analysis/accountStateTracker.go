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
	Normal StakeClaimType = 1

	// Transaction had market events.
	// NFT trading: staking NFT into market contract, claiming bought NFT,
	// claiming NFT after order cancelation, staking fungible tokens during bid,
	// claiming fungible tokens if bid was overbid.
	// TODO verify that above description is complete
	MarketEvent StakeClaimType = 2

	// Transaction was involved in SM rewards distribution.
	SmReward StakeClaimType = 3
)

func (t StakeClaimType) String() string {
	stakeClaimType := " "
	if t == MarketEvent {
		stakeClaimType = "M"
	} else if t == SmReward {
		stakeClaimType = "S"
	}
	return stakeClaimType
}

type AccountState struct {
	Tx             response.TransactionResult
	State          response.AccountResult
	IsSm           bool
	SmStateChanged bool
	StakeClaimType StakeClaimType
}

func CheckIfSmStateChanged(staked1, staked2 *big.Float) bool {
	return (staked1.Cmp(big.NewFloat(SmThreshold)) >= 0 && staked2.Cmp(big.NewFloat(SmThreshold)) < 0) ||
		(staked2.Cmp(big.NewFloat(SmThreshold)) >= 0 && staked1.Cmp(big.NewFloat(SmThreshold)) < 0)
}

func CheckIfSm(a *response.AccountResult) bool {
	return a.Stakes.ConvertDecimalsToFloat().Cmp(big.NewFloat(SmThreshold)) >= 0
}

func (a *AccountState) CheckIfSm() bool {
	return CheckIfSm(&a.State)
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
	processingDirection ProcessingDirection, tx string, stakeClaimType StakeClaimType) bool {

	smStateChanged := false

	eventKind := event.Unknown
	eventKind.SetString(e.Kind)

	var eventData *event.TokenEventData
	if !eventKind.IsTokenEvent() {
		return smStateChanged
	}

	// Decode event data into event.TokenEventData structure
	decoded, _ := hex.DecodeString(e.Data)
	eventData = io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

	if e.Address != a.Address {
		return smStateChanged
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
						originalStakedAmount := a.Stakes.ConvertDecimalsToFloat()

						a.Stake = amountAdd(currentSoulStaked, eventData.Value, processingDirection).String()
						a.Stakes.Amount = a.Stake

						newStakedAmount := a.Stakes.ConvertDecimalsToFloat()
						smStateChanged = CheckIfSmStateChanged(originalStakedAmount, newStakedAmount)
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
					originalStakedAmount := a.Stakes.ConvertDecimalsToFloat()

					a.Stake = amountSub(currentSoulStaked, eventData.Value, processingDirection).String()
					a.Stakes.Amount = a.Stake

					newStakedAmount := a.Stakes.ConvertDecimalsToFloat()
					smStateChanged = CheckIfSmStateChanged(originalStakedAmount, newStakedAmount)
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

		// NFTs are staked when they are listed on a marketplace
		case event.TokenStake:
			tokenBalance.Amount = amountSub(currentAmount, big.NewInt(1), processingDirection).String()
			idRemove(&tokenBalance.Ids, eventData.Value.String(), processingDirection)

		// CROWNS are claimed by eligible accounts during inflation transaction
		// NFTs are claimed when listing gets cancelled on a marketplace
		case event.TokenClaim:
			tokenBalance.Amount = amountAdd(currentAmount, big.NewInt(1), processingDirection).String()
			idAdd(&tokenBalance.Ids, eventData.Value.String(), processingDirection)
		}
	}

	return smStateChanged
}

func applyEventsToAccountState(es []response.EventResult, a *response.AccountResult, processingDirection ProcessingDirection, tx string) (bool, StakeClaimType) {
	stakeClaimType := Normal
	smStateChanged := false
	for _, e := range es {
		eventKind := event.Unknown
		eventKind.SetString(e.Kind)

		if eventKind == event.TokenMint {
			// Decode event data into event.TokenEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData := io.Deserialize[*event.TokenEventData](decoded, &event.TokenEventData{})

			if eventData.Symbol == "SOUL" {
				// TODO needs better detection but for now should work
				stakeClaimType = SmReward
				break
			}
		}

		if eventKind.IsMarketEvent() {
			// Decode event data into event.MarketEventData structure
			decoded, _ := hex.DecodeString(e.Data)
			eventData := io.Deserialize[*event.MarketEventData](decoded, &event.MarketEventData{})

			if eventData.QuoteSymbol == "SOUL" {
				stakeClaimType = MarketEvent
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

		if applyEventToAccountState(e,
			previousEvent,
			a,
			processingDirection, tx, stakeClaimType) {

			smStateChanged = true
		}

	}

	return smStateChanged, stakeClaimType
}

func applyTransactionToAccountState(tx response.TransactionResult, a *response.AccountResult, processingDirection ProcessingDirection) (bool, StakeClaimType) {
	// Skip failed trasactions
	if !tx.StateIsSuccess() {
		// State din't change, saving previous one
		return false, Normal
	}

	return applyEventsToAccountState(tx.Events, a, processingDirection, tx.Hash)
}

func resetUntrackedFields(account *response.AccountResult) {
	// We don't need them here, this array is not updated during state tracking atm
	account.Stakes.Unclaimed = ""
	account.Stakes.Time = 0
	account.Unclaimed = ""

	account.Validator = ""

	account.Storage = response.StorageResult{}
	account.Txs = nil
}

// TODO Work in progress
// txs should be ordered from first tx to last
// processingDirection == Forward: We are moving from first tx to last
// processingDirection == Backward: We are moving from last tx to first
// TrackAccountStateByEvents: Modifies account to the latest state using events in transactions, also returns account state array for each transaction
func TrackAccountStateByEvents(txs []response.TransactionResult,
	account *response.AccountResult, processingDirection ProcessingDirection) []AccountState {

	length := len(txs)
	state := make([]AccountState, length, length)

	resetUntrackedFields(account)

	for i := range txs {
		var txIndex int
		if processingDirection == Forward {
			txIndex = i
		} else {
			txIndex = length - 1 - i
		}

		state[txIndex].Tx = txs[txIndex]
		if processingDirection == Forward {
			// Modifying state first, saving it later, because processingDirection is Forward
			state[txIndex].SmStateChanged, state[txIndex].StakeClaimType = applyTransactionToAccountState(txs[txIndex], account, processingDirection)
			state[txIndex].State = *account.Clone()
			// Detecting if account is an SM in this state
			state[txIndex].IsSm = CheckIfSm(account)
		} else {
			// Saving state first, modifying it later, because processingDirection is Backward
			state[txIndex].State = *account.Clone()
			// Detecting if account is an SM in this state
			state[txIndex].IsSm = CheckIfSm(account)

			state[txIndex].SmStateChanged, state[txIndex].StakeClaimType = applyTransactionToAccountState(txs[txIndex], account, processingDirection)
		}
	}

	return state
}
