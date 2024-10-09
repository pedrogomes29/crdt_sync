package prolly

import "encoding/hex"

type KVStore[T any] map[string]T

func (kvStore KVStore[T]) Put(address [32]byte, value T) {
	kvStore[hex.EncodeToString(address[:])] = value
}

func (kvStore KVStore[T]) Delete(address [32]byte) {
	delete(kvStore, hex.EncodeToString(address[:]))
}

func (kvStore KVStore[T]) Get(address [32]byte) T {
	return kvStore[hex.EncodeToString(address[:])]
}
