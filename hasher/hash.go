package hasher

import (
	"crypto/sha256"
	"encoding/binary"
)

type Hasher interface {
	Hash() []byte
}

type Hint int
type Hstring string

func (x Hint) Hash() []byte {
	hasher := sha256.New()

	binaryInt := make([]byte, 4)
	binary.BigEndian.PutUint32(binaryInt, uint32(x))

	hasher.Write(binaryInt)

	return hasher.Sum(nil)
}

func (h Hstring) Hash() []byte {
	hasher := sha256.New()
	hasher.Write([]byte(h))
	return hasher.Sum(nil)
}
