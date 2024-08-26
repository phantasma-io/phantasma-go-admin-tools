package main

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/linxGnu/grocksdb"
)

type ListKeysWithUnknownBaseKeysVisitor struct {
	knownBaseKeysBytes map[BaseKey][]byte
}

func (v *ListKeysWithUnknownBaseKeysVisitor) Visit(it *grocksdb.Iterator) bool {
	key := it.Key()

	isKnown := false
	for _, b := range v.knownBaseKeysBytes {
		if bytes.HasPrefix(key.Data(), b) {
			isKnown = true
			break
		}
	}
	if !isKnown {
		fmt.Println(string(key.Data()))
	}

	key.Free()
	return true
}

type ListKeysWithUnknownSubKeysVisitor struct {
	baseKey      []byte
	knownSubKeys map[string][]byte
}

func (v *ListKeysWithUnknownSubKeysVisitor) Visit(it *grocksdb.Iterator) bool {
	key := it.Key()

	if !bytes.HasPrefix(key.Data(), v.baseKey) {
		key.Free()
		return true
	}

	isKnown := false
	for _, b := range v.knownSubKeys {
		if bytes.HasPrefix(key.Data(), b) {
			isKnown = true
			break
		}
	}
	if !isKnown {
		fmt.Println(string(key.Data()))
	}

	key.Free()
	return true
}

type ListUniqueSubKeysVisitor struct {
	baseKey      []byte
	FoundSubKeys [][]byte
	OverallFound int
}

func (v *ListUniqueSubKeysVisitor) Visit(it *grocksdb.Iterator) bool {
	key := it.Key()

	if !bytes.HasPrefix(key.Data(), v.baseKey) {
		key.Free()
		return true
	}

	v.OverallFound++

	subkeySrc := key.Data()[len(v.baseKey):]
	subkey := make([]byte, len(subkeySrc))
	copy(subkey, subkeySrc)

	i := slices.IndexFunc(v.FoundSubKeys, func(b []byte) bool {
		return bytes.Compare(b, subkey) == 0
	})

	if i == -1 {
		v.FoundSubKeys = append(v.FoundSubKeys, subkey)
		fmt.Println(string(subkey))
	}

	key.Free()
	return true
}

type LisContentsVisitor struct {
	KeyPrefix    []byte
	Limit        uint
	limitCounter uint
}

func (v *LisContentsVisitor) Visit(it *grocksdb.Iterator) bool {
	if v.Limit > 0 && v.limitCounter == v.Limit {
		return false
	}
	if v.Limit > 0 {
		v.limitCounter++
	}

	key := it.Key()

	if v.KeyPrefix != nil && !bytes.HasPrefix(key.Data(), v.KeyPrefix) {
		key.Free()
		return true
	}

	value := it.Value()
	//fmt.Printf("Key: %s Value: %v\n", string(key.Data()), value.Data())
	parsed, _ := ParseRow(key.Data(), value.Data())
	fmt.Println(parsed)

	key.Free()
	value.Free()

	return true
}
