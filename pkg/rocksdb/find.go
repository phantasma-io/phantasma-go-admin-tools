package rocksdb

import (
	"math/big"
	"slices"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
)

func (c *Connection) FindKeyByValue(value []byte) (bool, []byte) {
	it := c.db.NewIteratorCF(c.ro, c.cfHandles[1])

	it.SeekToFirst()

	for it = it; it.Valid(); it.Next() {
		valueSlice := it.Value()

		if slices.Equal(valueSlice.Data(), value) {
			keySlice := it.Key()
			key := keySlice.Data()
			keySlice.Free()
			return true, key
		}

		valueSlice.Free()
	}

	it.Close()
	return false, nil
}

func (c *Connection) FindElementIndex(keyPrefix, value []byte) (bool, *big.Int) {
	count, err := c.GetAsBigInt(storage.CountKey([]byte(keyPrefix)))
	if err != nil {
		panic(err)
	}

	var one = big.NewInt(1)
	for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
		key := storage.ElementKey(keyPrefix, i)
		v, err := c.Get(key)

		if err != nil {
			panic(err)
		}

		if slices.Equal(v, value) {
			return true, i
		}
	}

	return false, nil
}
