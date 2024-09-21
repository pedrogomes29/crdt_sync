package prolly

import "encoding/hex"

type KVStore map[string]interface{}

func (kvStore KVStore) Put(address [32]byte, value interface{}) {
	kvStore[hex.EncodeToString(address[:])] = value
}

func (kvStore KVStore) Get(address [32]byte) interface{} {
	return kvStore[hex.EncodeToString(address[:])]
}
