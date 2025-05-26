package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const stateFile = "state-progress.json"

type SimpleProgress struct {
	Data           map[string][]string
	ProcessedCount int
	EligibleTotal  int
	lock           sync.Mutex
}

// LoadProgress loads the JSON file from disk
func LoadProgress() (*SimpleProgress, error) {
	p := &SimpleProgress{
		Data: make(map[string][]string),
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

		arr, ok := v.([]any)
		if !ok {
			continue
		}

		months := make([]string, 0, len(arr))
		for _, m := range arr {
			if s, ok := m.(string); ok {
				months = append(months, s)
			}
		}
		p.Data[k] = months
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

	// Load original JSON from disk (merge-safe)
	raw := make(map[string]any)
	file, err := os.ReadFile(stateFile)
	if err == nil {
		_ = json.Unmarshal(file, &raw)
	}

	// Add the address
	raw[address] = months
	p.Data[address] = months

	// Update meta
	p.ProcessedCount++
	if len(months) > 0 {
		p.EligibleTotal++
	}
	raw["_meta"] = map[string]any{
		"processed":     p.ProcessedCount,
		"eligibleTotal": p.EligibleTotal,
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
