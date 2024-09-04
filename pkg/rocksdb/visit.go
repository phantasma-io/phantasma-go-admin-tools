package rocksdb

import (
	"github.com/linxGnu/grocksdb"
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
