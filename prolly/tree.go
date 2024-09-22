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
	return a[i].Less(a[j])
}

type KAddrPair[K hasher.Hasher, V hasher.Hasher] struct {
	key          K
	valueAddress [32]byte
}

func (p1 KAddrPair[K, V]) Less(p2 KAddrPair[K, V]) bool {
	p1KeyHash := p1.key.Hash()
	p2KeyHash := p2.key.Hash()
	return bytes.Compare(p1KeyHash[:], p2KeyHash[:]) == -1
}

func (p KAddrPair[K, V]) isLeaf(kvStore KVStore) bool {
	_, isInteriorNode := kvStore.Get(p.valueAddress).(ProllyTreeNode[K, V])
	return !isInteriorNode
}

func (p KAddrPair[K, V]) getProllyTreeNode(kvStore KVStore) ProllyTreeNode[K, V] {
	valueInterface := kvStore.Get(p.valueAddress)
	node, ok := valueInterface.(ProllyTreeNode[K, V])
	if !ok {
		panic(fmt.Sprintf("Type unknown: %T\n", valueInterface))
	}
	return node
}

func (p KAddrPair[K, V]) getKVPair(kvStore KVStore) KVPair[K, V] {
	valueInterface := kvStore.Get(p.valueAddress)
	value, ok := valueInterface.(V)
	if !ok {
		panic(fmt.Sprintf("Type unknown: %T\n", valueInterface))
	}
	return KVPair[K, V]{
		p.key,
		value,
	}
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

func (t1 ProllyTree[K, V]) Diff(t2 ProllyTree[K, V]) ([]KVPair[K, V], []KVPair[K, V]) {
	if t1.rootAddress == t2.rootAddress {
		return []KVPair[K, V]{}, []KVPair[K, V]{}
	}

	t1Root := t1.kvStore.Get(t1.rootAddress).(ProllyTreeNode[K, V])
	t2Root := t1.kvStore.Get(t2.rootAddress).(ProllyTreeNode[K, V])

	return FindNodesDiff([]ProllyTreeNode[K, V]{t1Root}, []ProllyTreeNode[K, V]{t2Root}, t1.kvStore, t2.kvStore)
}

func GetAllLeafs[K hasher.Hasher, V hasher.Hasher](nodes []ProllyTreeNode[K, V], kvStore KVStore) []ProllyTreeNode[K, V] {
	nodesAreLeafs := nodes[0].children[0].isLeaf(kvStore)
	if nodesAreLeafs {
		return nodes
	}
	var children []ProllyTreeNode[K, V]

	for _, node := range nodes {
		for _, child := range node.children {
			children = append(children, child.getProllyTreeNode(kvStore))
		}
	}

	return children
}

func FindNonMatchingPairs[K hasher.Hasher, V hasher.Hasher](t1Nodes []ProllyTreeNode[K, V], t2Nodes []ProllyTreeNode[K, V], t1KvStore KVStore, t2KvStore KVStore) (t1ExceptT2Pairs []KAddrPair[K, V], t2ExceptT1Pairs []KAddrPair[K, V]) {
	t1NodeIdx, t2NodeIdx := 0, 0
	t1NodeChildIdx, t2NodeChildIdx := 0, 0
	for t1NodeIdx < len(t1Nodes) && t2NodeIdx < len(t2Nodes) {
		t1Node := t1Nodes[t1NodeIdx]
		t2Node := t2Nodes[t2NodeIdx]
		for t1NodeChildIdx < len(t1Nodes) && t2NodeChildIdx < len(t2Nodes) {
			t1NodeChild := t1Node.children[t1NodeChildIdx]
			t2NodeChild := t2Node.children[t2NodeChildIdx]

			if t1NodeChild.key.Hash() == t2NodeChild.key.Hash() {
				if t1NodeChild.valueAddress != t2NodeChild.valueAddress {
					t1ExceptT2Pairs = append(t1ExceptT2Pairs, t1NodeChild)
					t2ExceptT1Pairs = append(t2ExceptT1Pairs, t2NodeChild)
				}
				t1NodeChildIdx++
				t2NodeChildIdx++
			} else if t1NodeChild.Less(t2NodeChild) {
				t1ExceptT2Pairs = append(t1ExceptT2Pairs, t1NodeChild)
				t1NodeChildIdx++
			} else {
				t2ExceptT1Pairs = append(t2ExceptT1Pairs, t2NodeChild)
				t2NodeChildIdx++
			}
		}
		if t1NodeChildIdx == len(t1Node.children) {
			t1NodeChildIdx = 0
			t1NodeIdx++
		}
		if t2NodeChildIdx == len(t2Node.children) {
			t2NodeChildIdx = 0
			t2NodeIdx++
		}
	}
	//TODO: handle different number of nodes and children per nodes

	return
}

func FindNodesDiff[K hasher.Hasher, V hasher.Hasher](t1Nodes []ProllyTreeNode[K, V], t2Nodes []ProllyTreeNode[K, V], t1KvStore KVStore, t2KvStore KVStore) ([]KVPair[K, V], []KVPair[K, V]) {
	t1NodesAreLeafs := t1Nodes[0].children[0].isLeaf(t1KvStore)
	t2NodesAreLeafs := t2Nodes[0].children[0].isLeaf(t2KvStore)
	if t1NodesAreLeafs && t2NodesAreLeafs {
		t1DiffKAddrPairs, t2DiffKAddrPairs := FindNonMatchingPairs(t1Nodes, t2Nodes, t1KvStore, t2KvStore)
		var t1ExceptT2pairs []KVPair[K, V]
		var t2ExceptT1Pairs []KVPair[K, V]

		for _, t1DiffKAddrPair := range t1DiffKAddrPairs {
			t1ExceptT2pairs = append(t1ExceptT2pairs, t1DiffKAddrPair.getKVPair(t1KvStore))
		}
		for _, t2DiffKAddrPair := range t2DiffKAddrPairs {
			t2ExceptT1Pairs = append(t2ExceptT1Pairs, t2DiffKAddrPair.getKVPair(t2KvStore))
		}

		return t1ExceptT2pairs, t2ExceptT1Pairs
	}

	var newT1Nodes []ProllyTreeNode[K, V]
	var newT2Nodes []ProllyTreeNode[K, V]
	if !t1NodesAreLeafs && !t2NodesAreLeafs {
		newT1NodesAddrPairs, newT2NodesAddrPairs := FindNonMatchingPairs(t1Nodes, t2Nodes, t1KvStore, t2KvStore)
		for _, newT1NodesAddrPair := range newT1NodesAddrPairs {
			newT1Nodes = append(newT1Nodes, newT1NodesAddrPair.getProllyTreeNode(t1KvStore))
		}
		for _, newT2NodesAddrPair := range newT2NodesAddrPairs {
			newT2Nodes = append(newT2Nodes, newT2NodesAddrPair.getProllyTreeNode(t2KvStore))
		}
	} else if !t1NodesAreLeafs {
		newT1Nodes = GetAllLeafs(t1Nodes, t1KvStore)
		newT2Nodes = t2Nodes
	} else if !t2NodesAreLeafs {
		newT1Nodes = t1Nodes
		newT2Nodes = GetAllLeafs(t1Nodes, t1KvStore)
	}

	return FindNodesDiff(newT1Nodes, newT2Nodes, t1KvStore, t2KvStore)
}
