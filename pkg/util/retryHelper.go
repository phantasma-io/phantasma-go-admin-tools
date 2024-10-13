package util

import (
	"fmt"
	"time"
)

func RetryHelper[T any](fn func() (T, error), maxRetry int, startBackoff, maxBackoff time.Duration) (T, error) {

	for attempt := 0; ; attempt++ {
		result, err := fn()
		if err == nil {
			return result, err
		}

		if attempt == maxRetry-1 {
			return result, err
		}

		fmt.Printf("Retrying after %s\n", startBackoff)
		time.Sleep(startBackoff)
		if maxBackoff == 0 || startBackoff < maxBackoff {
			startBackoff *= 2
		}
	}
}
