package main

import (
	"crypto/sha256"
	"encoding/binary"
)

type Hasher interface {
	Hash() []byte
}

type hint int
type hstring string

func (x hint) Hash() []byte {
	hasher := sha256.New()

	binaryInt := make([]byte, 4)
	binary.BigEndian.PutUint32(binaryInt, uint32(x))

	hasher.Write(binaryInt)

	return hasher.Sum(nil)
}

func (h hstring) Hash() []byte {
	hasher := sha256.New()
	hasher.Write([]byte(h))
	return hasher.Sum(nil)
}
