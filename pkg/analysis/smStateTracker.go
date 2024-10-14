package analysis

import (
	"fmt"
	"time"
)

const SmThreshold = 50000

// Processing direction is from current time to the past
func checkSmStateChangesDuringMonth(state []AccountState, year, month int, startOfNextMonthSmState bool) (bool, bool) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	// fmt.Printf("Checking interval %d - %d\n", start.UTC().Unix(), end.UTC().Unix())

	smStateChanged := false
	startOfThisMonthSmState := startOfNextMonthSmState
	for _, s := range state {
		if s.Tx.Timestamp > uint(start.UTC().Unix()) && s.Tx.Timestamp < uint(end.UTC().Unix()) {

			if s.SmStateChanged {
				smStateChanged = true
			}

			startOfThisMonthSmState = s.IsSm
		}
	}

	return startOfThisMonthSmState, smStateChanged
}

func DetectEligibleSm(currentSmState bool, states []AccountState, startingDate int64) []string {
	currentTime := time.Now().UTC()
	t := time.Unix(int64(startingDate), 0).UTC()
	startingYear := t.Year()
	startingMonth := int(t.Month())

	eligibleMonths := []string{}

	y := currentTime.Year()
	m := int(currentTime.Month())

	isEligibleSm := false

	isSmAtStartOfThisMonth, smStateChanged := checkSmStateChangesDuringMonth(states, y, m, currentSmState)

	for {
		m -= 1
		if m == 0 {
			m = 12
			y -= 1
		}

		isSmAtStartOfThisMonth, smStateChanged = checkSmStateChangesDuringMonth(states, y, m, isSmAtStartOfThisMonth)

		if smStateChanged {
			// State changed during month - not eligible
			isEligibleSm = false
		} else if isSmAtStartOfThisMonth {
			// State didn't changed during month & was SM - eligible
			isEligibleSm = true
		}

		if isEligibleSm {
			eligibleMonths = append(eligibleMonths, fmt.Sprintf("%d-%d", y, m))
		}
		// fmt.Printf("%d-%d: Setting this: isSmAtStartOfThisMonth: %t isEligibleSm: %t\n\n", y, m, isSmAtStartOfThisMonth, isEligibleSm)

		if y == startingYear && m == startingMonth {
			break
		}
	}

	return eligibleMonths
}
