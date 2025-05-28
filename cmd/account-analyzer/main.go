package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go/pkg/domain/event"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

var clients []rpc.PhantasmaRPC
var client rpc.PhantasmaRPC

var appOpts struct {
	Nexus                    string `short:"n" long:"nexus" description:"testnet or mainnet"`
	Order                    string `long:"order" default:"asc" description:"asc or desc"`
	ordering                 analysis.OrderDirection
	Output                   string   `short:"o" long:"output" description:"Output folder"`
	Input                    string   `short:"i" long:"input" description:"Data to process"`
	Offset                   uint     `long:"offset" description:"Offset for processed data"`
	BlockCache               string   `long:"block-cache" description:"Path to folder containing blocks cache"`
	Address                  string   `short:"a" long:"address" description:"Address to analyse"`
	AddressCsvPath           string   `long:"address-csv-path" description:"Path to CSV file with addresses"`
	IgnoreAddressCsvPath     string   `long:"ignore-address-csv-path" description:"Path to CSV file with addresses to ignore"`
	InvalidAddressOutputPath string   `long:"invalid-address-output-path" description:"Path to file to write invalid addresses to"`
	ErrorsOutputPath         string   `long:"errors-output-path" description:"Path to file to write errors to"`
	TokenSymbol              string   `long:"symbol" description:"Token symbol to track balance"`
	EventKinds               []string `long:"event-kind" description:"Filter out transactions which do not have these events"`
	ShowFungible             bool     `long:"show-fungible" description:"Show fungible token events and balances"`
	ShowNonfungible          bool     `long:"show-nonfungible" description:"Show nonfungible token events and balances"`
	ShowFailedTxes           bool     `long:"show-failed" description:"Shows failed transactions"`
	GetInitialState          bool     `long:"get-initial-state" description:"Get initial state of address by replaying transactions in reverse order"`
	GetSmStates              bool     `long:"get-sm-states" description:"Get per month SM states of address by replaying transactions in reverse order"`
	GetAllBlocks             bool     `long:"get-all-blocks" description:"Get all chain blocks"`
	GetKnownAddresses        bool     `long:"get-known-addresses" description:"Get all known addresses"`
	AnalyzeTxes              bool     `long:"analyze-txes" description:"Runs analysis of all transactions from cached blocks"`
	AnalyzeScript            bool     `long:"analyze-script" description:"Runs analysis of specified script"`
	GetRelatedAddresses      bool     `long:"get-related-addresses" description:"Get addresses which interacted with provided address"`
	TrackAccountState        bool     `long:"track-account-state" description:"Shows balance state of address for every displayed transaction"`
	UseInitialState          bool     `long:"use-initial-state" description:"Use initial state of address while replaying transactions with track-account-state argument"`
	ExportSmRewardsJson      string   `long:"export-sm-json" description:"Path to output JSON file with SM rewards"`
	Verbose                  bool     `short:"v" long:"verbose" description:"Verbose mode"`
}

var errorLogMutex sync.Mutex

func LogErrorToFile(filename, message string) {
	errorLogMutex.Lock()
	defer errorLogMutex.Unlock()

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, _ = f.WriteString(message + "\n")
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
		if appOpts.InvalidAddressOutputPath == "" {
			panic("--invalid-address-output-path argument must be set")
		}
		if appOpts.ErrorsOutputPath == "" {
			panic("--errors-output-path argument must be set")
		}

		addressesToIgnore := []string{}
		addresses := []string{}

		if appOpts.IgnoreAddressCsvPath != "" {
			f, err := os.Open(appOpts.IgnoreAddressCsvPath)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			csvReader := csv.NewReader(f)
			data, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			for _, row := range data {
				if len(row) > 0 {
					addressesToIgnore = append(addressesToIgnore, row[0])
				}
			}
		}

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

			for _, row := range data {
				if len(row) > 0 {
					addresses = append(addresses, row[0])
				}
			}
		} else {
			addresses = []string{appOpts.Address}
		}

		progress, err := LoadProgress()
		if err != nil {
			panic(fmt.Errorf("failed to load state: %w", err))
		}
		filtered := make([]string, 0, len(addresses))
		for _, addr := range addresses {
			if !progress.IsDone(addr) && !slices.Contains(addressesToIgnore, addr) {
				filtered = append(filtered, addr)
			}
		}
		addresses = filtered

		// 1669852800 - Thu Dec 01 2022 00:00:00 GMT+0000
		// 1701388800 - Fri Dec 01 2023 00:00:00 GMT+0000
		// 1704067200 - 2024-01-01 00:00:00 UTC
		startDate := int64(1704067200)

		if appOpts.Verbose {
			fmt.Printf("Processing %d addresses for SM rewards eligibility starting %s\n", len(addresses), time.Unix(startDate, 0).UTC().Format(time.RFC822))
		}

		if appOpts.ExportSmRewardsJson != "" {
			type MonthEntry struct {
				Count     int                `json:"count"`
				Payload   string             `json:"payload"`
				Addresses map[string]float64 `json:"addresses"`
			}

			const rewardPerMonth = 125000.0
			rewardsByMonth := make(map[string]*MonthEntry)

			numWorkers := 4
			sem := make(chan struct{}, numWorkers)
			var wg sync.WaitGroup
			var mu sync.Mutex

			notReported := 0
			start := time.Now()
			startIteration := time.Now()
			const reportEveryNIterations = 10

			for _, addr := range addresses {
				wg.Add(1)
				sem <- struct{}{}

				go func(addr string) {
					defer wg.Done()
					defer func() { <-sem }()

					months, errString := printSmStates(addr, startDate, appOpts.Verbose)
					if errString != "" {
						if strings.Contains(errString, "Address is invalid") {
							LogErrorToFile(appOpts.InvalidAddressOutputPath, addr)
						} else {
							LogErrorToFile(appOpts.ErrorsOutputPath, errString)
						}
						return
					}
					progress.SaveResult(addr, months)

					if appOpts.Verbose {
						mu.Lock()
						notReported++
						if notReported >= reportEveryNIterations {
							elapsed := time.Since(startIteration)
							fmt.Printf("Processed %d addresses [%d] in %.2f seconds, %.2f addresses per second\n",
								notReported, len(addresses), elapsed.Seconds(), float64(notReported)/elapsed.Seconds())
							notReported = 0
							startIteration = time.Now()
						}
						mu.Unlock()
					}
				}(addr)
			}

			wg.Wait()

			if appOpts.Verbose {
				fmt.Printf("Processed %d addresses in %f minutes\n", len(addresses), time.Since(start).Minutes())
			}

			// Build rewardsByMonth from progress state after all processing is done
			for addr, months := range progress.Data {
				for _, month := range months {
					entry := rewardsByMonth[month]
					if entry == nil {
						entry = &MonthEntry{
							Payload:   fmt.Sprintf("SM rewards for %s", month),
							Addresses: make(map[string]float64),
						}
						rewardsByMonth[month] = entry
					}
					entry.Addresses[addr] = 0 // placeholder, will be replaced
				}
			}

			// Fill actual rewards based on number of addresses per month
			for _, entry := range rewardsByMonth {
				count := len(entry.Addresses)
				if count > 0 {
					entry.Count = count
					reward := rewardPerMonth / float64(count)
					for k := range entry.Addresses {
						entry.Addresses[k] = reward
					}
				}
			}

			keys := make([]string, 0, len(rewardsByMonth))
			for k := range rewardsByMonth {
				keys = append(keys, k)
			}
			slices.Sort(keys)

			ordered := make(map[string]*MonthEntry)
			for _, k := range keys {
				ordered[k] = rewardsByMonth[k]
			}

			f, err := os.Create(appOpts.ExportSmRewardsJson)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			err = enc.Encode(ordered)
			if err != nil {
				panic(err)
			}
		} else {
			wr := csv.NewWriter(os.Stdout)

			for _, addr := range addresses {
				months, errString := printSmStates(addr, startDate, appOpts.Verbose)
				if errString != "" {
					panic(errString)
				}
				if len(months) > 0 {
					wr.Write(append([]string{addr}, months...))
					wr.Flush()
				}
			}
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
