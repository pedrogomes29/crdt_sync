package hasher

import (
	"crypto/sha256"
	"encoding/binary"
)

type Hasher interface {
	Hash(seed ...byte) [32]byte
}

type Hint int
type Hstring string

func (x Hint) Hash(seed ...byte) [32]byte {
	binaryInt := make([]byte, 4)
	binary.BigEndian.PutUint32(binaryInt, uint32(x))

	data := append(
		seed,
		binaryInt...,
	)
	return sha256.Sum256(data)
}

func (h Hstring) Hash(seed ...byte) [32]byte {
	data := append(
		seed,
		[]byte(h)...,
	)
	return sha256.Sum256(data)
}
