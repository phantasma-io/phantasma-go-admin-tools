package main

import (
	"bytes"
	"fmt"

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
