package analysis

import (
	"fmt"
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const SmThreshold = 50000

// Processing direction is from current time to the past
func checkSmStateChangesDuringMonth(perTxAccountBalances []AccountState, year, month int, startOfNextMonthSmState bool) (bool, bool) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	// fmt.Printf("Checking interval %d - %d\n", start.UTC().Unix(), end.UTC().Unix())

	smStateChanged := false
	startOfThisMonthSmState := startOfNextMonthSmState
	for _, s := range perTxAccountBalances {
		if s.Tx.Timestamp > uint(start.UTC().Unix()) && s.Tx.Timestamp < uint(end.UTC().Unix()) {

			if s.SmStateChanged {
				smStateChanged = true
			}

			startOfThisMonthSmState = s.IsSm
		}
	}

	return startOfThisMonthSmState, smStateChanged
}

func DetectEligibleSm(currentSmState bool, perTxAccountBalances []AccountState, startingDate int64) *orderedmap.OrderedMap[string, bool] {
	currentTime := time.Now().UTC()
	t := time.Unix(int64(startingDate), 0).UTC()
	startingYear := t.Year()
	startingMonth := int(t.Month())

	perMonthStates := orderedmap.New[string, bool]()

	y := currentTime.Year()
	m := int(currentTime.Month())

	isEligibleSm := false
	perMonthStates.Set(fmt.Sprintf("%d-%d", y, m), isEligibleSm)

	isSmAtStartOfThisMonth, smStateChanged := checkSmStateChangesDuringMonth(perTxAccountBalances, y, m, currentSmState)

	for {
		m -= 1
		if m == 0 {
			m = 12
			y -= 1
		}

		isSmAtStartOfThisMonth, smStateChanged = checkSmStateChangesDuringMonth(perTxAccountBalances, y, m, isSmAtStartOfThisMonth)

		if smStateChanged {
			// State changed during month - not eligible
			isEligibleSm = false
		} else if isSmAtStartOfThisMonth {
			// State didn't changed during month & was SM - eligible
			isEligibleSm = true
		}

		perMonthStates.Set(fmt.Sprintf("%d-%d", y, m), isEligibleSm)
		// fmt.Printf("%d-%d: Setting this: isSmAtStartOfThisMonth: %t isEligibleSm: %t\n\n", y, m, isSmAtStartOfThisMonth, isEligibleSm)

		if y == startingYear && m == startingMonth {
			break
		}
	}

	return perMonthStates
}