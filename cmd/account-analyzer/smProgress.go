package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

const stateFile = "state-progress.json"

type SimpleProgress struct {
	Data           map[string][]string
	Timestamps     map[string]int64 // new: stores last updated timestamps
	ProcessedCount int
	EligibleTotal  int
	lock           sync.Mutex
}

// LoadProgress loads the JSON file from disk
func LoadProgress() (*SimpleProgress, error) {
	p := &SimpleProgress{
		Data:       make(map[string][]string),
		Timestamps: make(map[string]int64),
	}

	bytes, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil // fresh
		}
		return nil, fmt.Errorf("cannot read %s: %w", stateFile, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", stateFile, err)
	}

	for k, v := range raw {
		if k == "_meta" {
			meta, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if val, ok := meta["processed"].(float64); ok {
				p.ProcessedCount = int(val)
			}
			if val, ok := meta["eligibleTotal"].(float64); ok {
				p.EligibleTotal = int(val)
			}
			continue
		}

		// BEGIN: v1 format parser (to be deleted later)
		if arr, ok := v.([]any); ok {
			months := make([]string, 0, len(arr))
			for _, m := range arr {
				if s, ok := m.(string); ok {
					months = append(months, s)
				}
			}
			p.Data[k] = months
			p.Timestamps[k] = time.Now().Unix() // new: backfill timestamp
			continue
		}
		// END: v1 format parser

		// v2 format
		if entry, ok := v.(map[string]any); ok {
			months := []string{} // always initialize as empty slice
			if arr, ok := entry["months"].([]any); ok {
				for _, m := range arr {
					if s, ok := m.(string); ok {
						months = append(months, s)
					}
				}
			}
			p.Data[k] = months

			if ts, ok := entry["updated"].(float64); ok {
				p.Timestamps[k] = int64(ts)
			} else {
				p.Timestamps[k] = time.Now().Unix() // new: backfill if missing
			}
		}
	}

	return p, nil
}

// IsDone returns true if address was already recorded
func (p *SimpleProgress) IsDone(address string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, ok := p.Data[address]
	return ok
}

// SaveResult records new result and saves the full file to disk
func (p *SimpleProgress) SaveResult(address string, months []string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// No changes if already exists
	if _, ok := p.Data[address]; ok {
		return
	}

	// Ensure months is at least an empty slice to avoid null in JSON
	if months == nil {
		months = []string{}
	}

	// Add the address
	p.Data[address] = months
	p.Timestamps[address] = time.Now().Unix() // new: save timestamp

	// Update meta
	p.ProcessedCount++
	if len(months) > 0 {
		p.EligibleTotal++
	}

	raw := make(map[string]any, len(p.Data)+1)
	raw["_meta"] = map[string]any{
		"processed":     p.ProcessedCount,
		"eligibleTotal": p.EligibleTotal,
	}

	for addr, months := range p.Data {
		raw[addr] = map[string]any{
			"months":  months,
			"updated": p.Timestamps[addr], // new: include timestamp in output
		}
	}

	// Save back
	bytes, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal state: %v\n", err)
		return
	}
	if err := os.WriteFile(stateFile, bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write state: %v\n", err)
	}
}
