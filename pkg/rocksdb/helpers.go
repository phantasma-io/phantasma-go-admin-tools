package rocksdb

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func PrepareOpts() *grocksdb.Options {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	return opts
}

func ListColumnFamilies(dbPath string) {
	opts := PrepareOpts()

	columnFamilyNames, err := grocksdb.ListColumnFamilies(opts, dbPath)
	if err != nil {
		panic(err.Error())
	}

	for _, f := range columnFamilyNames {
		fmt.Println(f)
	}
}
