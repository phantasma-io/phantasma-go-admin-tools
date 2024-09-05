package storage

import "math/big"

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

type Balance struct {
	TokenSymbol string
	Address     string
	Amount      *big.Int
}

func (b Balance) String() string {
	return "TokenSymbol: " + b.TokenSymbol + " Address: " + b.Address + " Amount: " + b.Amount.String()
}
