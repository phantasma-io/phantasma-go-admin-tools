package storage

import (
	"encoding/json"
	"strings"
)

type KeyValue struct {
	Key   string
	Value string
}

func (k KeyValue) String() string {
	return "Key: " + k.Key + " Value: " + k.Value
}

func (k KeyValue) ToSlice() []string {
	return []string{k.Value}
}

type KeyValueJson struct {
	Key   string
	Value any
}

func (k KeyValueJson) String() string {
	j, err := json.Marshal(k.Value)
	if err != nil {
		panic(err)
	}
	return "Key: " + k.Key + " Value: " + string(j)
}

func (k KeyValueJson) ToSlice() []string {
	j, err := json.Marshal(k.Value)
	if err != nil {
		panic(err)
	}
	return []string{string(j)}
}

type Address struct {
	Address string
	Name    string
}

func (a Address) String() string {
	return "Address: " + a.Address + " Name: " + a.Name
}

func (a Address) ToSlice() []string {
	return []string{a.Address, a.Name}
}

type BalanceFungible struct {
	TokenSymbol string
	Address     string
	Amount      string
}

func (b BalanceFungible) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Amount: " + b.Amount
}

func (b BalanceFungible) ToSlice() []string {
	return []string{b.TokenSymbol, b.Address, b.Amount}
}

type BalanceNonFungibleSingleRow struct {
	TokenSymbol string
	Address     string
	Id          string
}

func (b BalanceNonFungibleSingleRow) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Ids: " + b.Id
}

func (b BalanceNonFungibleSingleRow) ToSlice() []string {
	return []string{b.TokenSymbol, b.Address, b.Id}
}

// Converting []BalanceNonFungibleSingleRow to []*BalanceNonFungible, grouping balances by addresses
func ConvertBalanceNonFungibleFromSingleRows(singleRowBalances []any) []any {
	result := make([]any, 0)

	for _, sAny := range singleRowBalances {

		s := sAny.(BalanceNonFungibleSingleRow)

		var targetBalance any
		for _, r := range result {
			if r.(*BalanceNonFungible).TokenSymbol == s.TokenSymbol && r.(*BalanceNonFungible).Address == s.Address {
				targetBalance = r
			}
		}

		if targetBalance == nil {
			targetBalance = &BalanceNonFungible{TokenSymbol: s.TokenSymbol, Address: s.Address, Ids: []string{}}
			result = append(result, targetBalance)
		}

		targetBalance.(*BalanceNonFungible).Ids = append(targetBalance.(*BalanceNonFungible).Ids, s.Id)
	}

	return result
}

type BalanceNonFungible struct {
	TokenSymbol string
	Address     string
	Ids         []string
}

func (b BalanceNonFungible) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Ids: " + strings.Join(b.Ids, " ")
}

type Tx struct {
	TxHash       string
	TxHashB64    string
	BlockHashB64 string
	BlockHeight  string
	TxBytesB64   string
}

func (t Tx) String() string {
	return "TxHash: " + t.TxHash + " BlockHash: " + t.BlockHashB64 + " BlockHeight: " + t.BlockHeight + " TxBytes: " + t.TxBytesB64
}
func (t Tx) ToSlice() []string {
	panic("not supported")
}

type BlockHeightAndHash struct {
	Height  string
	Hash    string
	HashB64 string
}

func (b BlockHeightAndHash) String() string {
	return "Height: " + b.Height + " Hash: " + b.Hash + " HashB64: " + b.HashB64
}
func (b BlockHeightAndHash) ToSlice() []string {
	panic("not supported")
}

type Block struct {
	Height    string
	Hash      string
	Timestamp uint32
	Bytes     string
}

func (b Block) String() string {
	return "Height: " + b.Height + " Hash: " + b.Hash + " Bytes: " + b.Bytes
}
func (b Block) ToSlice() []string {
	panic("not supported")
}
