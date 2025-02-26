package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

type BlockHeight struct {
	Height  string
	HashB64 []byte
}

var appOpts struct {
	DbPath                      string `short:"p" long:"db-path" description:"Path to Rocksdb database directory" required:"true"`
	ColumnFamily                string `short:"f" long:"column-family" description:"Column family to open"`
	ListColumnFamilies          bool   `long:"list-column-families" description:"Lists column families available in the database"`
	DumpData                    bool   `short:"d" long:"dump" description:"Dump data of given column family"`
	DumpAddresses               bool   `long:"dump-addresses" description:"Dump all addresses"`
	DumpTokenSymbols            bool   `long:"dump-token-symbols" description:"Dump token symbols of all fungible and non-fungible tokens"`
	DumpTokenInfo               bool   `long:"dump-token-info" description:"Dump token information"`
	DumpBalances                bool   `long:"dump-balances" description:"Dump balances of all fungible tokens for all addresses"`
	DumpBalancesNft             bool   `long:"dump-balances-nft" description:"Dump balances of all non-fungible tokens for all addresses"`
	DumpBlockHashes             bool   `long:"dump-block-hashes" description:"Dump all block hashes"`
	DumpBlocks                  bool   `long:"dump-blocks" description:"Dump all blocks"`
	DumpTransactions            bool   `long:"dump-txes" description:"Dump all transactions"`
	DumpStakingClaims           bool   `long:"dump-staking-claims" description:"Dump staking claims"`
	DumpStakingLeftovers        bool   `long:"dump-staking-leftovers" description:"Dump staking KCAL leftovers (part of unclaimed)"`
	DumpStakingMasterAge        bool   `long:"dump-staking-master-age" description:"Dump staking master age map"`
	DumpStakingMasterClaims     bool   `long:"dump-staking-master-claims" description:"Dump staking master claims timestamps"`
	DumpStakes                  bool   `long:"dump-stakes" description:"Dump stakes"`
	DumpNfts                    bool   `long:"dump-nfts" description:"Dump nfts"`
	DumpSeries                  bool   `long:"dump-series" description:"Dump nft series"`
	DumpContractNames           bool   `long:"dump-contract-names" description:"Dump names of deployed contracts"`
	DumpContractInfos           bool   `long:"dump-contract-infos" description:"Dump common information about deployed contracts"`
	MergeKcalLeftovers          bool   `long:"merge-kcal-leftovers" description:"Merge KCAL leftovers to balances"`
	Decompress                  bool   `long:"decompress" description:"Decompress blocks and txes, works with --dump-blocks and --dump-txes. False by default"`
	BaseKey                     string `long:"base-key" description:"Filter contents by base key"`
	SubKeys                     string `long:"subkeys" description:"Subkeys for given base key which needs to be dumped (coma-separated)"`
	SubKeysCsv                  string `long:"subkeys-csv" description:"Subkeys for given base key which needs to be dumped (path to csv file)"`
	SubKeysCsv2                 string `long:"subkeys-csv2" description:"Subkeys for given base key which needs to be dumped (path to csv file), will be merged with subkeys-csv"`
	Addresses                   string `long:"addresses" description:"Addresses for filtering out results"`
	BlockHeightsJson            string `long:"block-heigts-json" description:"JSON with block heights and hashes, result of --dump-block-hashes"`
	FungibleBalancesJson        string `long:"fungible-balances-json" description:"JSON with fungible balances, result of --dump-balances"`
	NftBalancesJson             string `long:"nft-balances-json" description:"JSON with nft balances, result of --dump-balances-nft"`
	KcalLeftoversJson           string `long:"kcal-leftovers-json" description:"JSON with KCAL leftovers, result of --dump-staking-leftovers"`
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

	subKeysSlice     []string
	blockHeightsMap  map[string]int
	blockHeightsMap2 map[int]string
	fungibleBalances []storage.BalanceFungible
	nftBalances      []storage.BalanceNonFungible
	kcalLeftovers    []storage.KeyValue
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

		if appOpts.SubKeysCsv2 != "" {
			// We parse CSV file into appOpts.SubKeys slice

			f, err := os.Open(appOpts.SubKeysCsv2)
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

	if appOpts.DumpStakingClaims || appOpts.DumpStakes || appOpts.DumpStakingLeftovers || appOpts.DumpStakingMasterAge {
		if len(appOpts.subKeysSlice) == 0 {
			panic("This argument requires addresses passed with --subkeys-csv or --subkeys keys")
		}
	}

	if appOpts.DumpTransactions || appOpts.DumpBlocks {
		if len(appOpts.BlockHeightsJson) == 0 {
			panic("This argument requires block height JSON file path to be provided with --block-heigts-json")
		}

		f, err := os.Open(appOpts.BlockHeightsJson)
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		b, _ := io.ReadAll(f)
		var parsed []BlockHeight
		json.Unmarshal(b, &parsed)

		if appOpts.DumpTransactions {
			appOpts.blockHeightsMap = make(map[string]int)
			for i, p := range parsed {
				appOpts.blockHeightsMap[string(p.HashB64)] = i
			}
		} else if appOpts.DumpBlocks {
			appOpts.blockHeightsMap2 = make(map[int]string)
			for i, p := range parsed {
				appOpts.blockHeightsMap2[i] = string(p.HashB64)
			}
		}

		if appOpts.Verbose {
			fmt.Printf("Loaded %d block heigts and hashes", len(appOpts.blockHeightsMap))
		}
	}

	if appOpts.DumpNfts || appOpts.DumpSeries {
		if len(appOpts.NftBalancesJson) == 0 {
			panic("This argument requires nft balances JSON file path to be provided with --nft-balances-json")
		}

		f, err := os.Open(appOpts.NftBalancesJson)
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		b, _ := io.ReadAll(f)
		json.Unmarshal(b, &appOpts.nftBalances)
	}
	if appOpts.MergeKcalLeftovers {
		{
			if len(appOpts.FungibleBalancesJson) == 0 {
				panic("This argument requires fungible balances JSON file path to be provided with --fungible-balances-json")
			}

			f, err := os.Open(appOpts.FungibleBalancesJson)
			if err != nil {
				fmt.Println(err)
			}
			defer f.Close()

			b, _ := io.ReadAll(f)
			json.Unmarshal(b, &appOpts.fungibleBalances)
		}

		{
			if len(appOpts.KcalLeftoversJson) == 0 {
				panic("This argument requires KCAL leftovers JSON file path to be provided with --dump-staking-leftovers")
			}

			f, err := os.Open(appOpts.KcalLeftoversJson)
			if err != nil {
				fmt.Println(err)
			}
			defer f.Close()

			b, _ := io.ReadAll(f)
			json.Unmarshal(b, &appOpts.kcalLeftovers)
		}
	}

	if appOpts.DumpData || appOpts.DumpAddresses || appOpts.DumpTokenSymbols || appOpts.DumpBalances || appOpts.DumpBalancesNft ||
		appOpts.DumpBlockHashes || appOpts.DumpBlocks || appOpts.DumpTransactions ||
		appOpts.DumpStakingClaims || appOpts.DumpStakes ||
		appOpts.DumpStakingLeftovers || appOpts.DumpStakingMasterAge || appOpts.DumpStakingMasterClaims ||
		appOpts.DumpNfts || appOpts.DumpSeries {
		dump()
	} else if appOpts.DumpTokenInfo {
		dump_TokenInfo()
	}

	if appOpts.MergeKcalLeftovers {
		AddKcalLeftovers()
	}

	if appOpts.DumpContractNames {
		dump_ContractsNames()
	}
	if appOpts.DumpContractInfos {
		dump_ContractsInfos()
	}
}
