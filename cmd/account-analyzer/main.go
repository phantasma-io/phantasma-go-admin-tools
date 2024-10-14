package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

var clients []rpc.PhantasmaRPC
var client rpc.PhantasmaRPC

var appOpts struct {
	Nexus             string `short:"n" long:"nexus" description:"testnet or mainnet"`
	Order             string `long:"order" default:"asc" description:"asc or desc"`
	ordering          analysis.OrderDirection
	Output            string   `short:"o" long:"output" description:"Output folder"`
	BlockCache        string   `long:"block-cache" description:"Path to folder containing blocks cache"`
	Address           string   `short:"a" long:"address" description:"Address to analyse"`
	TokenSymbol       string   `long:"symbol" description:"Token symbol to track balance"`
	EventKinds        []string `long:"event-kind" description:"Filter out transactions which do not have these events"`
	ShowFungible      bool     `long:"show-fungible" description:"Show fungible token events and balances"`
	ShowNonfungible   bool     `long:"show-nonfungible" description:"Show nonfungible token events and balances"`
	ShowFailedTxes    bool     `long:"show-failed" description:"Shows failed transactions"`
	GetInitialState   bool     `long:"get-initial-state" description:"Get initial state of address by replaying transactions in reverse order"`
	GetSmStates       bool     `long:"get-sm-states" description:"Get per month SM states of address by replaying transactions in reverse order"`
	GetAllBlocks      bool     `long:"get-all-blocks" description:"Get all chain blocks"`
	GetKnownAddresses bool     `long:"get-known-addresses" description:"Get all known addresses"`
	TrackAccountState bool     `long:"track-account-state" description:"Shows balance state of address for every displayed transaction"`
	UseInitialState   bool     `long:"use-initial-state" description:"Use initial state of address while replaying transactions with track-account-state argument"`
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

	if appOpts.Nexus == "testnet" {
		client = rpc.NewRPCTestnet()
		clients = []rpc.PhantasmaRPC{client}
	} else {
		clients = rpc.NewRPCSetMainnet()
		client = clients[0]
	}
	analysis.InitChainTokens(client)

	cfgSymbol = appOpts.TokenSymbol

	for _, karg := range appOpts.EventKinds {
		k := event.Unknown
		k.SetString(karg)

		cfgEventKinds = append(cfgEventKinds, k)
	}

	cfgShowFungible = appOpts.ShowFungible
	cfgShowNonfungible = appOpts.ShowNonfungible

	cfgShowFailedTxes = appOpts.ShowFailedTxes

	if appOpts.GetInitialState {
		printOriginalState(appOpts.Address)
	} else if appOpts.GetSmStates {
		// 1669852800 - Thu Dec 01 2022 00:00:00 GMT+0000
		// 1701388800 - Fri Dec 01 2023 00:00:00 GMT+0000
		printSmStates(appOpts.Address, 1669852800)
	} else if appOpts.GetKnownAddresses {
		addresses := analysis.GetAllKnownAddresses(clients, appOpts.BlockCache)

		for _, r := range addresses {
			fmt.Printf("%s,", r)
		}
	} else if appOpts.GetAllBlocks {
		if appOpts.Output == "" {
			panic("--output argument is mandatory when --get-all-blocks passed")
		}

		analysis.GetAllBlocks(appOpts.Output, clients)
	} else {
		printTransactions(appOpts.Address, appOpts.TrackAccountState, appOpts.UseInitialState, appOpts.ordering)
	}
}
