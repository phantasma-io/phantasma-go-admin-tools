package rocksdb

import (
	"github.com/linxGnu/grocksdb"
)

type Connection struct {
	db        *grocksdb.DB
	cfHandles []*grocksdb.ColumnFamilyHandle
	ro        *grocksdb.ReadOptions
	err       error
}

func NewConnection(dbPath, columnFamily string) *Connection {
	c := Connection{}
	c.Connect(dbPath, columnFamily)

	return &c
}

func (c *Connection) Connect(dbPath, columnFamily string) error {
	opts := PrepareOpts()

	var err error
	c.db, c.cfHandles, err = grocksdb.OpenDbForReadOnlyColumnFamilies(opts, dbPath, []string{"default", columnFamily}, []*grocksdb.Options{opts, opts /*grocksdb.NewDefaultOptions()*/}, false)
	if err != nil {
		return err
	}

	c.ro = grocksdb.NewDefaultReadOptions()
	return nil
}

func (c *Connection) Destroy() {

	for _, h := range c.cfHandles {
		h.Destroy()
	}
	c.cfHandles = nil

	c.db.Close()
	c.db = nil
}
