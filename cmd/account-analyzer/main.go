package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

var client rpc.PhantasmaRPC

var appOpts struct {
	Nexus           string `short:"n" long:"nexus" description:"testnet or mainnet"`
	Order           string `long:"order" default:"asc" description:"asc or desc"`
	ordering        analysis.OrderDirection
	Address         string   `short:"a" long:"address" description:"Address to analyse"`
	TokenSymbol     string   `long:"symbol" description:"Token symbol to track balance"`
	EventKinds      []string `long:"event-kind" description:"Filter out transactions which do not have these events"`
	ShowFungible    bool     `long:"show-fungible" description:"Show fungible token events and balances"`
	ShowNonfungible bool     `long:"show-nonfungible" description:"Show nonfungible token events and balances"`
	ShowFailedTxes  bool     `long:"show-failed" description:"Shows failed transactions"`
	GetInitialState bool     `long:"get-initial-state" description:"Get initial state of address by replaying transactions in reverse order"`
	GetStakingTxes  bool     `long:"get-staking-txes" description:"Get staking transaction hashes for address"`
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
	} else {
		client = rpc.NewRPCMainnet()
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
	} else if appOpts.GetStakingTxes {
		cfgEventKinds = []event.EventKind{event.TokenStake}
		printTransactions(appOpts.Address, false, appOpts.ordering)
	} else {
		printTransactions(appOpts.Address, true, appOpts.ordering)
	}
}
