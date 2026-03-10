package rocksdb

import (
	"bytes"

	"github.com/linxGnu/grocksdb"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

type Visitor interface {
	Visit(it *grocksdb.Iterator) bool
}

func (c *Connection) Visit(visitor Visitor) {
	it := c.db.NewIteratorCF(c.ro, c.cfHandles[1])

	it.SeekToFirst()

	for it = it; it.Valid(); it.Next() {
		if !visitor.Visit(it) {
			break
		}
	}

	it.Close()
}

func (c *Connection) VisitPrefix(prefix []byte, visitor func(key []byte, value []byte) bool) {
	it := c.db.NewIteratorCF(c.ro, c.cfHandles[1])
	it.Seek(prefix)

	for ; it.Valid(); it.Next() {
		keySlice := it.Key()
		if !bytes.HasPrefix(keySlice.Data(), prefix) {
			keySlice.Free()
			break
		}

		valueSlice := it.Value()
		key := util.ArrayClone(keySlice.Data())
		value := util.ArrayClone(valueSlice.Data())
		keySlice.Free()
		valueSlice.Free()

		if !visitor(key, value) {
			break
		}
	}

	it.Close()
}
