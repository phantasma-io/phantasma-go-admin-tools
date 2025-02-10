package main

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func AddKcalLeftovers() {
	o := NewOutput(OutputFormatFromString(appOpts.OutputFormat))

	totalAmountAdded := big.NewInt(0)

	count := 0

	for _, leftover := range appOpts.kcalLeftovers {
		address := leftover.Key
		amount := leftover.Value

		for i, b := range appOpts.fungibleBalances {
			if b.Address == address && b.TokenSymbol == "KCAL" {
				originalBalance, ok := big.NewInt(0).SetString(b.Amount, 10)
				if !ok {
					panic("Cannot parse amount")
				}

				leftoverAmount, ok := big.NewInt(0).SetString(amount, 10)
				if !ok {
					panic("Cannot parse leftover amount")
				}

				newBalance, ok := big.NewInt(0).SetString(b.Amount, 10)
				if !ok {
					panic("Cannot parse amount")
				}
				newBalance = newBalance.Add(originalBalance, leftoverAmount)

				totalAmountAdded = totalAmountAdded.Add(totalAmountAdded, leftoverAmount)

				count++
				fmt.Println(strconv.Itoa(count) + " " + address + " " + b.Amount + " -> " + newBalance.String() + " [" + util.ConvertDecimals(originalBalance, 10) + " + " + util.ConvertDecimals(leftoverAmount, 10) + " -> " + util.ConvertDecimals(newBalance, 10) + "]")

				b.Amount = newBalance.String()
				appOpts.fungibleBalances[i] = b
			}
		}
	}

	fmt.Println("Total amount of KCAL added: " + util.ConvertDecimals(totalAmountAdded, 10))

	for _, b := range appOpts.fungibleBalances {
		o.AddJsonRecord(b)
	}

	o.Flush()
}
