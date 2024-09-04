package rocksdb

import (
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func (c *Connection) Get(key []byte) ([]byte, error) {
	rdbSlice, err := c.db.GetCF(c.ro, c.cfHandles[1], key)
	if err != nil {
		return nil, err
	}

	result := util.ArrayClone(rdbSlice.Data())
	rdbSlice.Free()

	return result, err
}

func (c *Connection) GetAsBigInt(key []byte) (*big.Int, error) {
	b, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	return util.BigIntFromCsharpOrPhantasmaByteArray(b), err
}
