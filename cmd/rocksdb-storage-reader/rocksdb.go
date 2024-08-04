package main

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func RocksdbPrepareOpts() *grocksdb.Options {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	return opts
}

func RocksdbListColumnFamilies(dbPath string) {
	opts := RocksdbPrepareOpts()

	columnFamilyNames, err := grocksdb.ListColumnFamilies(opts, dbPath)
	if err != nil {
		panic(err.Error())
	}

	for _, f := range columnFamilyNames {
		fmt.Println(f)
	}
}

type Visitor interface {
	Visit(it *grocksdb.Iterator) bool
}

func RocksdbDbRoVisit(dbPath, columnFamily string, visitor Visitor) {
	opts := RocksdbPrepareOpts()

	db, columnFamilyHandlers, err := grocksdb.OpenDbForReadOnlyColumnFamilies(opts, dbPath, []string{"default", columnFamily}, []*grocksdb.Options{opts, opts /*grocksdb.NewDefaultOptions()*/}, false)
	if err != nil {
		panic(err.Error())
	}

	ro := grocksdb.NewDefaultReadOptions()

	it := db.NewIteratorCF(ro, columnFamilyHandlers[1])

	it.SeekToFirst()

	for it = it; it.Valid(); it.Next() {
		if !visitor.Visit(it) {
			break
		}
	}

	it.Close()

	for _, h := range columnFamilyHandlers {
		h.Destroy()
	}

	db.Close()
}
