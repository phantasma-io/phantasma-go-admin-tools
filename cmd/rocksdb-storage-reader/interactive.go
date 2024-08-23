package main

import (
	"bytes"
	"fmt"

	"github.com/linxGnu/grocksdb"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/console"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-go/pkg/io"
)

type ListBalancesForAddressVisitor struct {
	keys   [][]byte
	labels []string
}

func (v *ListBalancesForAddressVisitor) Visit(it *grocksdb.Iterator) bool {
	key := it.Key()

	for i, k := range v.keys {
		if bytes.Compare(key.Data(), k) == 0 {
			value := it.Value()

			br := *io.NewBinReaderFromBuf(value.Data())
			number := br.ReadBigInteger()
			value.Free()

			fmt.Println(v.labels[i] + ": " + number.String())
		}
	}

	key.Free()
	return true
}

func listBalancesForAddress(addressStr string) {
	var address cryptography.Address
	address, err := cryptography.FromString(addressStr)
	if err != nil {
		panic(err)
	}
	// We must use Bytes() here to get address without length byte prefix (34, '"')
	addressBytes := address.Bytes()

	keys := make([][]byte, len(KnowSubKeys[Balances]))
	for i, t := range KnowSubKeys[Balances] {
		keys[i] = GetBalanceTokenAddressKey(addressBytes, t)
	}

	v := ListBalancesForAddressVisitor{keys: keys, labels: KnowSubKeys[Balances]}
	RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)
}

type GetNameForAddressVisitor struct {
	key []byte
}

func (v *GetNameForAddressVisitor) Visit(it *grocksdb.Iterator) bool {
	key := it.Key()

	if bytes.Compare(key.Data(), v.key) == 0 {
		value := it.Value()

		name := string(value.Data())
		value.Free()

		fmt.Println("Name: " + name)
		return false
	}

	key.Free()
	return true
}

func getNameForAddress(addressStr string) {
	var address cryptography.Address
	address, err := cryptography.FromString(addressStr)
	if err != nil {
		panic(err)
	}
	// We must use Serialize() here to get additional length byte prefix (34, '"')
	addressBytes := io.Serialize[*cryptography.Address](&address)

	v := GetNameForAddressVisitor{key: GetAccountAddressMapKey(addressBytes)}
	RocksdbDbRoVisit(appOpts.DbPath, appOpts.ColumnFamily, &v)
}

func interactiveMainMenu() {
	logout := false
	for !logout {
		menuIndex, _ := console.PromptIndexedMenu("\nROCKSDB STORAGE VIEWER. MENU:",
			[]string{"Get balances for address",
				"Get name for address",
				"Logout"})

		switch menuIndex {
		case 1:
			address := console.PromptStringInput("Enter address:")
			listBalancesForAddress(address)
		case 2:
			address := console.PromptStringInput("Enter address:")
			getNameForAddress(address)
		case 3:
			logout = true
		}
	}
}
