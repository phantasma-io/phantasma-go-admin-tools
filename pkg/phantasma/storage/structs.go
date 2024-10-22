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

type Address struct {
	Address string
	Name    string
}

func (a Address) String() string {
	return "Address: " + a.Address + " Name: " + a.Name
}

type BalanceFungible struct {
	TokenSymbol string
	Address     string
	Amount      *big.Int
}

func (b BalanceFungible) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Amount: " + b.Amount.String()
}

type BalanceNonFungible struct {
	TokenSymbol string
	Address     string
	Ids         []string
}

func (b BalanceNonFungible) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Ids: " + strings.Join(b.Ids, " ")
}
