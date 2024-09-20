package prolly

import (
	"crdt_sync/hasher"
	"crypto/sha256"
)

type KVPair[K hasher.Hasher, V hasher.Hasher] struct{
	key K;
	value V;
}

type ProllyTree[K hasher.Hasher, V hasher.Hasher] struct {
	children []KVPair[K,V];
}


func (pt ProllyTree[K, V]) Hash() []byte {
	hasher := sha256.New()

	for _,kvPair := range pt.children{
		hasher.Write(kvPair.key.Hash())
		hasher.Write(kvPair.value.Hash())
	}

	return hasher.Sum(nil)
}
