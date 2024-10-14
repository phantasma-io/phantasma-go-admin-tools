package main

import (
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
)

func printAllKnownAddresses() {
	addresses := analysis.GetAllKnownAddresses(clients)

	for _, r := range addresses {
		fmt.Printf("%s,", r)
	}
}
