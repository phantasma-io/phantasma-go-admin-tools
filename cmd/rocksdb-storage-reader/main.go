package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/rocksdb"
)

var appOpts struct {
	DbPath                      string `short:"p" long:"db-path" description:"Path to Rocksdb database directory" required:"true"`
	ColumnFamily                string `short:"f" long:"column-family" description:"Column family to open"`
	ListColumnFamilies          bool   `long:"list-column-families" description:"Lists column families available in the database"`
	DumpData                    bool   `short:"d" long:"dump" description:"Dump data of given column family"`
	BaseKey                     string `long:"base-key" description:"Filter contents by base key"`
	ListKeysWithUnknownBaseKeys bool   `long:"list-keys-with-unknown-base-keys" description:"Show keys with unknown base keys"`
	ListKeysWithUnknownSubKeys  bool   `long:"list-keys-with-unknown-sub-keys" description:"Show keys with unknown sub keys. base-key argument is mandatory if this flag is passed"`
	ListUniqueSubKeys           bool   `long:"list-unique-sub-keys" description:"Show unique sub keys for given base key. base-key argument is mandatory if this flag is passed"`
	ParseSubkeyAsAddress        bool   `long:"parse-subkey-as-address" description:"Try parsing subkeys as addresses"`
	ParseSubkeyAsHash           bool   `long:"parse-subkey-as-hash" description:"Try parsing subkeys as hashes"`
	OutputFormat                string `long:"output-format" description:"Format to use for data output: CSV, JSON, PLAIN"`
	Limit                       uint   `long:"limit" description:"Limit processing with given amount of rows"`
	Interactive                 bool   `short:"i" long:"interactive" description:"Interactive mode"`
	Verbose                     bool   `short:"v" long:"verbose" description:"Verbose mode"`
}

func main() {
	_, err := flags.Parse(&appOpts)
	if err != nil {
		panic(err)
	}

	if appOpts.Interactive {
		interactiveMainMenu()
		return
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

	if appOpts.DumpData {
		dump()
	}
}
