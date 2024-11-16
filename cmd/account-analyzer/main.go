package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

var clients []rpc.PhantasmaRPC
var client rpc.PhantasmaRPC

var appOpts struct {
	Nexus               string `short:"n" long:"nexus" description:"testnet or mainnet"`
	Order               string `long:"order" default:"asc" description:"asc or desc"`
	ordering            analysis.OrderDirection
	Output              string   `short:"o" long:"output" description:"Output folder"`
	Input               string   `short:"i" long:"input" description:"Data to process"`
	Offset              uint     `long:"offset" description:"Offset for processed data"`
	BlockCache          string   `long:"block-cache" description:"Path to folder containing blocks cache"`
	Address             string   `short:"a" long:"address" description:"Address to analyse"`
	AddressCsvPath      string   `long:"address-csv-path" description:"Path to CSV file with addresses"`
	TokenSymbol         string   `long:"symbol" description:"Token symbol to track balance"`
	EventKinds          []string `long:"event-kind" description:"Filter out transactions which do not have these events"`
	ShowFungible        bool     `long:"show-fungible" description:"Show fungible token events and balances"`
	ShowNonfungible     bool     `long:"show-nonfungible" description:"Show nonfungible token events and balances"`
	ShowFailedTxes      bool     `long:"show-failed" description:"Shows failed transactions"`
	GetInitialState     bool     `long:"get-initial-state" description:"Get initial state of address by replaying transactions in reverse order"`
	GetSmStates         bool     `long:"get-sm-states" description:"Get per month SM states of address by replaying transactions in reverse order"`
	GetAllBlocks        bool     `long:"get-all-blocks" description:"Get all chain blocks"`
	GetKnownAddresses   bool     `long:"get-known-addresses" description:"Get all known addresses"`
	AnalyzeTxes         bool     `long:"analyze-txes" description:"Runs analysis of all transactions from cached blocks"`
	AnalyzeScript       bool     `long:"analyze-script" description:"Runs analysis of specified script"`
	GetRelatedAddresses bool     `long:"get-related-addresses" description:"Get addresses which interacted with provided address"`
	TrackAccountState   bool     `long:"track-account-state" description:"Shows balance state of address for every displayed transaction"`
	UseInitialState     bool     `long:"use-initial-state" description:"Use initial state of address while replaying transactions with track-account-state argument"`
	Verbose             bool     `short:"v" long:"verbose" description:"Verbose mode"`
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
		printOriginalState(appOpts.Address, appOpts.Verbose)
	} else if appOpts.GetSmStates {
		addresses := []string{}

		if appOpts.AddressCsvPath != "" {
			f, err := os.Open(appOpts.AddressCsvPath)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			csvReader := csv.NewReader(f)
			data, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			addresses = data[0]
		} else {
			addresses = []string{appOpts.Address}
		}

		// 1669852800 - Thu Dec 01 2022 00:00:00 GMT+0000
		// 1701388800 - Fri Dec 01 2023 00:00:00 GMT+0000
		startDate := int64(1701388800)

		if appOpts.Verbose {
			fmt.Printf("Processing %d addresses for SM rewards eligibility starting %s", len(addresses), time.Unix(startDate, 0).UTC().Format(time.RFC822))
		}

		notReported := 0
		start := time.Now()
		startIteration := time.Now()

		for _, a := range addresses {
			if appOpts.Verbose {
				const reportEveryNIterations = 100
				if notReported >= reportEveryNIterations {
					elapsed := time.Since(startIteration)
					fmt.Printf("Processed %d addresses [%d] in %f seconds, %f addresses per second\n", notReported, len(addresses), elapsed.Seconds(), float64(notReported)/elapsed.Seconds())
					notReported = 0
					startIteration = time.Now()
				}
			}

			printSmStates(a, startDate, appOpts.Verbose)

			notReported += 1
		}

		if appOpts.Verbose {
			fmt.Printf("Processed %d addresses in %f minutes\n", len(addresses), time.Since(start).Minutes())
		}
	} else if appOpts.GetKnownAddresses {
		addresses := analysis.GetAllKnownAddresses(clients, appOpts.BlockCache, appOpts.Verbose)

		for _, r := range addresses {
			fmt.Printf("%s,", r)
		}
	} else if appOpts.GetAllBlocks {
		if appOpts.Output == "" {
			panic("--output argument is mandatory when --get-all-blocks passed")
		}

		analysis.GetAllBlocks(appOpts.Output, clients)
	} else if appOpts.AnalyzeTxes {
		if appOpts.BlockCache == "" {
			panic("--block-cache argument is mandatory when --analyze-txes")
		}

		analysis.TxScriptsAnalyzer(appOpts.BlockCache, appOpts.Verbose, clients)
	} else if appOpts.AnalyzeScript {
		if appOpts.Input == "" {
			panic("--input argument is mandatory when --analyze-script")
		}

		calls := make(map[string]uint)
		analysis.ScriptsAnalyzer(appOpts.Input, appOpts.Offset, 18, appOpts.Verbose, calls, clients)

		for k, n := range calls {
			fmt.Printf("%s: %d\n", k, n)
		}
	} else {
		printTransactions(appOpts.Address, appOpts.TrackAccountState, appOpts.UseInitialState, appOpts.ordering, appOpts.Verbose, appOpts.GetRelatedAddresses)
	}
}
