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

func (k *KeyBuilder) AppendBytes(b []byte) {
	k.key = append(k.key, b...)
}

func (k *KeyBuilder) AppendString(s string) {
	k.key = append(k.key, []byte(s)...)
}

func (k *KeyBuilder) AppendAddress(address cryptography.Address) {
	k.key = append(k.key, address.Bytes()...)
}

func (k *KeyBuilder) AppendAddressAsString(address string) {
	a, _ := cryptography.FromString(address)
	k.AppendAddress(a)
}

func (k *KeyBuilder) Build() []byte {
	return k.key
}
