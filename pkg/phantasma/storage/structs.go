package storage

import (
	"math/big"
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
	Amount      *big.Int
}

func (b BalanceFungible) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Amount: " + b.Amount.String()
}

func (b BalanceFungible) ToSlice() []string {
	return []string{b.TokenSymbol, b.Address, b.Amount.String()}
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
