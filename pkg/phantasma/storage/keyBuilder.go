package storage

import (
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

type KeyBuilder struct {
	key []byte
}

func KeyBuilderNew() *KeyBuilder {

	return &KeyBuilder{key: []byte{}}
}

func (k *KeyBuilder) SetBytes(b []byte) *KeyBuilder {
	k.key = b
	return k
}

func (k *KeyBuilder) AppendBytes(b []byte) *KeyBuilder {
	k.key = append(k.key, b...)
	return k
}

func (k *KeyBuilder) AppendString(s string) *KeyBuilder {
	k.key = append(k.key, []byte(s)...)
	return k
}

func (k *KeyBuilder) AppendAddress(address cryptography.Address) *KeyBuilder {
	k.key = append(k.key, address.Bytes()...)
	return k
}

func (k *KeyBuilder) AppendAddressAsString(address string) *KeyBuilder {
	a, _ := cryptography.FromString(address)
	k.AppendAddress(a)
	return k
}

func (k *KeyBuilder) AppendAddressPrefixed(address cryptography.Address) *KeyBuilder {
	k.key = append(k.key, address.BytesPrefixed()...)
	return k
}

func (k *KeyBuilder) AppendAddressPrefixedAsString(address string) *KeyBuilder {
	a, _ := cryptography.FromString(address)
	k.AppendAddressPrefixed(a)
	return k
}

func (k *KeyBuilder) Build() []byte {
	return k.key
}
