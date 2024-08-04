package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/console"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

var client rpc.PhantasmaRPC

var appOpts struct {
	Nexus           string `short:"n" long:"nexus" description:"testnet or mainnet"`
	Order           string `long:"order" default:"asc" description:"asc or desc"`
	ordering        analysis.OrderDirection
	Address         string `short:"a" long:"address" description:"Address to analyse"`
	TokenSymbol     string `long:"symbol" description:"Token symbol to track balance"`
	ShowFungible    bool   `long:"show-fungible" description:"Show fungible token events and balances"`
	ShowNonfungible bool   `long:"show-nonfungible" description:"Show nonfungible token events and balances"`
	Interactive     bool   `short:"i" long:"interactive" description:"Interactive mode"`
}

func main() {
	_, err := flags.Parse(&appOpts)
	if err != nil {
		panic(err)
	}

	if appOpts.Order == "asc" {
		appOpts.ordering = analysis.Asc
	} else if appOpts.Order == "desc" {
		appOpts.ordering = analysis.Desc
	} else {
		panic("Unknown value of 'order' argument: " + appOpts.Order)
	}

	if appOpts.Interactive {
		if appOpts.Nexus == "" {
			_, appOpts.Nexus = console.PromptIndexedMenu("SELECT TESTNET OR MAINNET", []string{"testnet", "mainnet"})

			if appOpts.Nexus == "testnet" {
				client = rpc.NewRPCTestnet()
			} else {
				client = rpc.NewRPCMainnet()
			}
		}

		if appOpts.Address == "" {
			appOpts.Address = console.PromptStringInput("Enter address: ")
		}

		tokenCount := analysis.InitChainTokens(client)
		fmt.Println("Received information about", tokenCount, appOpts.Nexus, "tokens")

		interactiveMainMenu()
		return
	} else {
		if appOpts.Nexus == "testnet" {
			client = rpc.NewRPCTestnet()
		} else {
			client = rpc.NewRPCMainnet()
		}
		analysis.InitChainTokens(client)

		printTransactions(appOpts.Address, appOpts.TokenSymbol, appOpts.ordering, false, appOpts.ShowFungible, appOpts.ShowNonfungible)
	}
}
