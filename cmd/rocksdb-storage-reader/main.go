package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

var appOpts struct {
	DbPath                      string `short:"p" long:"db-path" description:"Path to Rocksdb database directory" required:"true"`
	ColumnFamily                string `short:"f" long:"column-family" description:"Column family to open"`
	ListColumnFamilies          bool   `long:"list-column-families" description:"Lists column families available in the database"`
	DumpData                    bool   `short:"d" long:"dump" description:"Dump data of given column family"`
	DumpAddresses               bool   `long:"dump-addresses" description:"Dump all addresses"`
	DumpTokenSymbols            bool   `long:"dump-token-symbols" description:"Dump token symbols of all fungible and non-fungible tokens"`
	DumpBalances                bool   `long:"dump-balances" description:"Dump balances of all fungible tokens for all addresses"`
	DumpBalancesNft             bool   `long:"dump-balances-nft" description:"Dump balances of all non-fungible tokens for all addresses"`
	DumpBlockHashes             bool   `long:"dump-block-hashes" description:"Dump all block hashes"`
	DumpBlocks                  bool   `long:"dump-blocks" description:"Dump all blocks"`
	DumpStakingClaims           bool   `long:"dump-staking-claims" description:"Dump staking claims"`
	DumpStakes                  bool   `long:"dump-stakes" description:"Dump stakes"`
	BaseKey                     string `long:"base-key" description:"Filter contents by base key"`
	SubKeys                     string `long:"subkeys" description:"Subkeys for given base key which needs to be dumped (coma-separated)"`
	SubKeysCsv                  string `long:"subkeys-csv" description:"Subkeys for given base key which needs to be dumped (path to csv file)"`
	Addresses                   string `long:"addresses" description:"Addresses for filtering out results"`
	PanicOnUnknownSubkey        bool   `long:"panic-on-unknown-subkey" description:"Crash if unknown subkey was detected"`
	ListKeysWithUnknownBaseKeys bool   `long:"list-keys-with-unknown-base-keys" description:"Show keys with unknown base keys"`
	ListKeysWithUnknownSubKeys  bool   `long:"list-keys-with-unknown-sub-keys" description:"Show keys with unknown sub keys. base-key argument is mandatory if this flag is passed"`
	ListUniqueSubKeys           bool   `long:"list-unique-sub-keys" description:"Show unique sub keys for given base key. base-key argument is mandatory if this flag is passed"`
	ParseSubkeyAsAddress        bool   `long:"parse-subkey-as-address" description:"Try parsing subkeys as addresses"`
	ParseSubkeyAsHash           bool   `long:"parse-subkey-as-hash" description:"Try parsing subkeys as hashes"`
	Output                      string `long:"output" description:"Output file path, if not set everything is printed into standard output"`
	OutputFormat                string `long:"output-format" description:"Format to use for data output: CSV, JSON, PLAIN"`
	Limit                       uint   `long:"limit" description:"Limit processing with given amount of rows"`
	Verbose                     bool   `short:"v" long:"verbose" description:"Verbose mode"`

	subKeysSlice []string
}

func main() {
	_, err := flags.Parse(&appOpts)
	if err != nil {
		panic(err)
	}

	if appOpts.SubKeysCsv != "" {
		// We parse CSV file into appOpts.SubKeys slice

		if appOpts.SubKeys != "" {
			panic("subkeys-csv and subkeys keys cannot be used at the same time")
		}

		f, err := os.Open(appOpts.SubKeysCsv)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		csvReader := csv.NewReader(f)
		data, err := csvReader.ReadAll()
		if err != nil {
			panic(err)
		}

		for _, k := range data {
			appOpts.subKeysSlice = append(appOpts.subKeysSlice, k[0])
		}
	} else if appOpts.SubKeys != "" {
		appOpts.subKeysSlice = strings.Split(appOpts.SubKeys, ",")
	}

	if appOpts.ListKeysWithUnknownBaseKeys {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)

		v := ListKeysWithUnknownBaseKeysVisitor{knownBaseKeysBytes: GetBytesForKnownBaseKeys()}
		c.Visit(&v)

		c.Destroy()
		return
	}
	if appOpts.ListKeysWithUnknownSubKeys {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)

		v := ListKeysWithUnknownSubKeysVisitor{baseKey: []byte(appOpts.BaseKey),
			knownSubKeys: GetBytesForKnownSubKeys(BaseKey(appOpts.BaseKey), true)}
		c.Visit(&v)

		c.Destroy()
		return
	}

	if appOpts.ListUniqueSubKeys {
		c := rocksdb.NewConnection(appOpts.DbPath, appOpts.ColumnFamily)

		v := ListUniqueSubKeysVisitor{baseKey: []byte(appOpts.BaseKey),
			FoundSubKeys: [][]byte{},
			OverallFound: 0}
		c.Visit(&v)

		c.Destroy()

		if appOpts.Verbose {
			fmt.Printf("Found %d unique keys out of %d keys overall\n", len(v.FoundSubKeys), v.OverallFound)
		}
		return
	}

	if appOpts.ListColumnFamilies {
		rocksdb.ListColumnFamilies(appOpts.DbPath)
		return
	}

	if appOpts.DumpStakingClaims || appOpts.DumpStakes {
		if len(appOpts.subKeysSlice) == 0 {
			panic("This argument requires addresses passed with --subkeys-csv or --subkeys keys")
		}
	}

	if appOpts.DumpData || appOpts.DumpAddresses || appOpts.DumpTokenSymbols || appOpts.DumpBalances || appOpts.DumpBalancesNft ||
		appOpts.DumpBlockHashes || appOpts.DumpBlocks || appOpts.DumpStakingClaims || appOpts.DumpStakes {
		dump()
	}
}
