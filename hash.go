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

	return hasher.Sum(nil)
}

func (h hstring) Hash() []byte {
	hf := sha256.New()
	hf.Write([]byte(h))
	return hf.Sum(nil)
}
