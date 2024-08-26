package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

var appOpts struct {
	DbPath                      string `short:"p" long:"db-path" description:"Path to Rocksdb database directory" required:"true"`
	ColumnFamily                string `short:"f" long:"column-family" description:"Column family to open"`
	ListColumnFamilies          bool   `long:"list-column-families" description:"Lists column families available in the database"`
	ListContents                bool   `short:"l" long:"list-contents" description:"Lists contents of given column family"`
	BaseKey                     string `long:"base-key" description:"Filter contents by base key"`
	ListKeysWithUnknownBaseKeys bool   `long:"list-keys-with-unknown-base-keys" description:"Show keys with unknown base keys"`
	ListKeysWithUnknownSubKeys  bool   `long:"list-keys-with-unknown-sub-keys" description:"Show keys with unknown sub keys. base-key argument is mandatory if this flag is passed"`
	ListUniqueSubKeys           bool   `long:"list-unique-sub-keys" description:"Show unique sub keys for given base key. base-key argument is mandatory if this flag is passed"`
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
		v := ListKeysWithUnknownBaseKeysVisitor{knownBaseKeysBytes: GetBytesForKnownBaseKeys()}
		RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)
		return
	}
	if appOpts.ListKeysWithUnknownSubKeys {
		v := ListKeysWithUnknownSubKeysVisitor{baseKey: []byte(appOpts.BaseKey),
			knownSubKeys: GetBytesForKnownSubKeys(BaseKey(appOpts.BaseKey), true)}
		RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)
		return
	}

	if appOpts.ListUniqueSubKeys {
		v := ListUniqueSubKeysVisitor{baseKey: []byte(appOpts.BaseKey),
			FoundSubKeys: [][]byte{},
			OverallFound: 0}
		RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)

		if appOpts.Verbose {
			fmt.Printf("Found %d unique keys out of %d keys overall\n", len(v.FoundSubKeys), v.OverallFound)
		}
		return
	}

	if appOpts.ListColumnFamilies {
		RocksdbListColumnFamilies(appOpts.DbPath)
		return
	}

	if appOpts.ListContents {
		v := LisContentsVisitor{Limit: appOpts.Limit}
		if appOpts.BaseKey != "" {
			v.KeyPrefix = []byte(appOpts.BaseKey)
		}
		RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)
	}
}
