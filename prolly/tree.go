package prolly

import (
	"bytes"
	"crdt_sync/hasher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

const TARGET_BITS = 2 //how many bits must be 0 to cause a new boundary => expected block size is 2^TARGET_BITS
var Target *big.Int

func init() {
	Target = big.NewInt(1)
	Target.Lsh(Target, uint(256-TARGET_BITS))
}

type ByKeyHash[K hasher.Hasher, V hasher.Hasher] []KAddrPair[K, V]

func (a ByKeyHash[K, V]) Len() int {
	return len(a)
}
func (a ByKeyHash[K, V]) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByKeyHash[K, V]) Less(i, j int) bool {
	pairIKeyHash := a[i].key.Hash()
	pairJKeyHash := a[j].key.Hash()
	return bytes.Compare(pairIKeyHash[:], pairJKeyHash[:]) == -1
}

type KAddrPair[K hasher.Hasher, V hasher.Hasher] struct {
	key          K
	valueAddress [32]byte
}

type KVPair[K hasher.Hasher, V hasher.Hasher] struct {
	key   K
	value V
}

type ProllyTreeNode[K hasher.Hasher, V hasher.Hasher] struct {
	children []KAddrPair[K, V]
}

type ProllyTree[K hasher.Hasher, V hasher.Hasher] struct {
	rootAddress [32]byte
	kvStore     KVStore
}

func (pair KAddrPair[K, V]) CheckBoundary(level hasher.Hint) bool {
	levelHash := level.Hash()
	keyHash := pair.key.Hash(levelHash[:]...)
	var hashInt big.Int
	hashInt.SetBytes(keyHash[:])

	isBoundary := hashInt.Cmp(Target) == -1
	return isBoundary
}

func indexValues[K hasher.Hasher, V hasher.Hasher](kvPairs []KVPair[K, V], kvStore KVStore) []KAddrPair[K, V] {
	var kAddrPairs []KAddrPair[K, V]
	for _, kvPair := range kvPairs {
		valueAddr := kvPair.value.Hash()
		kvStore.Put(valueAddr, kvPair.value)
		kAddrPairs = append(kAddrPairs, KAddrPair[K, V]{
			kvPair.key,
			valueAddr,
		})
	}
	return kAddrPairs
}

func chunkPairs[K hasher.Hasher, V hasher.Hasher](pairs []KAddrPair[K, V], level hasher.Hint, kvStore KVStore) []KAddrPair[K, V] {
	var newLevelPairs []KAddrPair[K, V]
	var currentChunk = ProllyTreeNode[K, V]{}
	for _, kAddrPair := range pairs {
		currentChunk.children = append(currentChunk.children, kAddrPair)
		if kAddrPair.CheckBoundary(level) {
			currentChunkAddr := currentChunk.Hash()
			newKAddrPair := KAddrPair[K, V]{
				kAddrPair.key,
				currentChunkAddr,
			}
			kvStore.Put(currentChunkAddr, currentChunk)
			newLevelPairs = append(newLevelPairs, newKAddrPair)
			currentChunk = ProllyTreeNode[K, V]{}
		}
	}

	if len(currentChunk.children) > 0 { //if last pair didn't cause a split
		highestKeyPair := currentChunk.children[len(currentChunk.children)-1]
		currentChunkAddr := currentChunk.Hash()
		newKAddrPair := KAddrPair[K, V]{
			highestKeyPair.key,
			currentChunkAddr,
		}
		kvStore.Put(currentChunkAddr, currentChunk)
		newLevelPairs = append(newLevelPairs, newKAddrPair)
	}

	return newLevelPairs
}

func InitKVPair[K hasher.Hasher, V hasher.Hasher](key K, value V) *KVPair[K, V] {
	return &KVPair[K, V]{
		key,
		value,
	}
}

func InitProllyTree[K hasher.Hasher, V hasher.Hasher](kvPairs []KVPair[K, V]) *ProllyTree[K, V] {
	tree := &ProllyTree[K, V]{
		kvStore: make(map[string]interface{}),
	}

	kAddrPairs := indexValues(kvPairs, tree.kvStore)
	sort.Sort(ByKeyHash[K, V](kAddrPairs))

	level := hasher.Hint(0)
	for len(kAddrPairs) > 1 {
		kAddrPairs = chunkPairs(kAddrPairs, level, tree.kvStore)
		level += 1
	}

	tree.rootAddress = kAddrPairs[0].valueAddress

	return tree
}

func (tree ProllyTree[K, V]) String() string {
	var sb strings.Builder
	var queue [][32]byte
	sb.WriteString("\nTree:\n")
	queue = append(queue, tree.rootAddress)
	for len(queue) > 0 { //bfs
		nodeCount := len(queue)
		for nodeCount > 0 { //visit all nodes at the current depth
			address := queue[0]
			queue = queue[1:]

			switch addressee := tree.kvStore.Get(address).(type) {
			case ProllyTreeNode[K, V]:
				for _, child := range addressee.children {
					keyHash := child.key.Hash()
					sb.WriteString(fmt.Sprintf("key hash: %s ", hex.EncodeToString(keyHash[:])))
					sb.WriteString(fmt.Sprintf("hash: %s  ", hex.EncodeToString(child.valueAddress[:])))
					queue = append(queue, child.valueAddress)
				}
			case V:
				sb.WriteString(fmt.Sprintf("%s", addressee))
			default:
				panic(fmt.Sprintf("Type unknown: %T\n", addressee))
			}
			sb.WriteString("\n")
			nodeCount--
		}

		sb.WriteString("\n\n")

	}
	return sb.String()
}

func (node ProllyTreeNode[K, V]) Hash(seed ...byte) [32]byte {
	data := seed

	for _, kvPair := range node.children {
		keyHash := kvPair.key.Hash()

		data = append(data, keyHash[:]...)
		data = append(data, kvPair.valueAddress[:]...)
	}

	return sha256.Sum256(data)
}
