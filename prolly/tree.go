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

type ByKeyHash[K hasher.Hasher] []KAddrPair[K]

func (a ByKeyHash[K]) Len() int {
	return len(a)
}
func (a ByKeyHash[K]) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByKeyHash[K]) Less(i, j int) bool {
	pairIKeyHash := a[i].key.Hash()
	pairJKeyHash := a[j].key.Hash()
	return bytes.Compare(pairIKeyHash[:], pairJKeyHash[:]) == -1
}

type KAddrPair[K hasher.Hasher] struct {
	key          K
	valueAddress [32]byte
}

type KVPair[K hasher.Hasher, V hasher.Hasher] struct {
	key   K
	value V
}

type ProllyTreeNode[K hasher.Hasher] struct {
	children []KAddrPair[K]
}

type ProllyTree[K hasher.Hasher, V hasher.Hasher] struct {
	rootAddress [32]byte
	kvStore     map[string]interface{}
}

func (k KAddrPair[K]) CheckBoundary(level hasher.Hint) bool {
	levelHash := level.Hash()
	keyHash := k.key.Hash(levelHash[:]...)
	var hashInt big.Int
	hashInt.SetBytes(keyHash[:])

	isBoundary := hashInt.Cmp(Target) == -1
	return isBoundary
}

func indexValues[K hasher.Hasher, V hasher.Hasher](kvPairs []KVPair[K, V], kvStore map[string]interface{}) []KAddrPair[K] {
	var kAddrPairs []KAddrPair[K]
	for _, kvPair := range kvPairs {
		valueAddr := kvPair.value.Hash()
		kvStore[hex.EncodeToString(valueAddr[:])] = kvPair.value
		kAddrPairs = append(kAddrPairs, KAddrPair[K]{
			kvPair.key,
			valueAddr,
		})
	}
	return kAddrPairs
}

func chunkPairs[K hasher.Hasher](pairs []KAddrPair[K], level hasher.Hint, kvStore map[string]interface{}) []KAddrPair[K] {
	var newLevelPairs []KAddrPair[K]
	var currentChunk = ProllyTreeNode[K]{}
	for _, kAddrPair := range pairs {
		currentChunk.children = append(currentChunk.children, kAddrPair)
		if kAddrPair.CheckBoundary(level) {
			currentChunkAddr := currentChunk.Hash()
			newKAddrPair := KAddrPair[K]{
				kAddrPair.key,
				currentChunkAddr,
			}
			kvStore[hex.EncodeToString(currentChunkAddr[:])] = currentChunk
			newLevelPairs = append(newLevelPairs, newKAddrPair)
			currentChunk = ProllyTreeNode[K]{}
		}
	}

	if len(currentChunk.children) > 0 { //if last pair didn't cause a split
		highestKeyPair := currentChunk.children[len(currentChunk.children)-1]
		currentChunkAddr := currentChunk.Hash()
		newKAddrPair := KAddrPair[K]{
			highestKeyPair.key,
			currentChunkAddr,
		}
		kvStore[hex.EncodeToString(currentChunkAddr[:])] = currentChunk
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
	sort.Sort(ByKeyHash[K](kAddrPairs))

	level := hasher.Hint(0)
	for len(kAddrPairs) > 1 {
		kAddrPairs = chunkPairs(kAddrPairs, level, tree.kvStore)
		level += 1
	}

	tree.rootAddress = kAddrPairs[0].valueAddress

	return tree
}

func (pt ProllyTree[K, V]) String() string {
	var sb strings.Builder
	var queue [][32]byte
	sb.WriteString("\nTree:\n")
	queue = append(queue, pt.rootAddress)
	for len(queue) > 0 {//bfs
		nodeCount := len(queue)
		for nodeCount > 0 { //visit all nodes at the current depth
			address := queue[0]
			queue = queue[1:]

			addressStr := hex.EncodeToString(address[:])

			switch addressee := pt.kvStore[addressStr].(type) {
			case ProllyTreeNode[K]:
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

func (pt ProllyTreeNode[K]) Hash(seed ...byte) [32]byte {
	data := seed

	for _, kvPair := range pt.children {
		keyHash := kvPair.key.Hash()

		data = append(data, keyHash[:]...)
		data = append(data, kvPair.valueAddress[:]...)
	}

	return sha256.Sum256(data)
}


func (pt ProllyTreeNode[K]) Hash(seed ...byte) [32]byte {
}