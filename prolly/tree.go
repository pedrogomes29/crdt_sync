package prolly

import (
	"bytes"
	"crdt_sync/hasher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
)

type ByKeyHash[K hasher.Hasher] []KAddrPair[K]

func (a ByKeyHash[K]) Len() int {
	return len(a)
}
func (a ByKeyHash[K]) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByKeyHash[K]) Less(i, j int) bool {
	return a[i].Less(a[j])
}

type KAddrPair[K hasher.Hasher] struct {
	key          K
	valueAddress [32]byte
}

func (p1 KAddrPair[K]) Less(p2 KAddrPair[K]) bool {
	p1KeyHash := p1.key.Hash()
	p2KeyHash := p2.key.Hash()
	return bytes.Compare(p1KeyHash[:], p2KeyHash[:]) == -1
}

func (p KAddrPair[K]) isLeaf() bool {
	return reflect.ValueOf(p.valueAddress).IsZero()
}

func (p KAddrPair[K]) getProllyTreeNode(kvStore KVStore) ProllyTreeNode[K] {
	valueInterface := kvStore.Get(p.valueAddress)
	node, ok := valueInterface.(ProllyTreeNode[K])
	if !ok {
		panic(fmt.Sprintf("Type unknown: %T\n", valueInterface))
	}
	return node
}

type ProllyTreeNode[K hasher.Hasher] struct {
	children []KAddrPair[K]
}

type ProllyTreeConfig struct {
	Target *big.Int
}

type ProllyTree[K hasher.Hasher] struct {
	rootAddress [32]byte
	kvStore     KVStore
	config      ProllyTreeConfig
}

func (pair KAddrPair[K]) CheckBoundary(level hasher.Hint, target *big.Int) bool {
	levelHash := level.Hash()
	keyHash := pair.key.Hash(levelHash[:]...)
	var hashInt big.Int
	hashInt.SetBytes(keyHash[:])

	isBoundary := hashInt.Cmp(target) == -1
	return isBoundary
}

func indexValues[K hasher.Hasher](keys []K) []KAddrPair[K] {
	var kAddrPairs []KAddrPair[K]
	for _, key := range keys {
		kAddrPairs = append(kAddrPairs, KAddrPair[K]{
			key: key,
		})
	}
	return kAddrPairs
}

func chunkPairs[K hasher.Hasher](pairs []KAddrPair[K], level hasher.Hint, kvStore KVStore, config ProllyTreeConfig) []KAddrPair[K] {
	var newLevelPairs []KAddrPair[K]
	var currentChunk = ProllyTreeNode[K]{}
	for _, kAddrPair := range pairs {
		currentChunk.children = append(currentChunk.children, kAddrPair)
		if kAddrPair.CheckBoundary(level, config.Target) {
			currentChunkAddr := currentChunk.Hash()
			newKAddrPair := KAddrPair[K]{
				kAddrPair.key,
				currentChunkAddr,
			}
			kvStore.Put(currentChunkAddr, currentChunk)
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
		kvStore.Put(currentChunkAddr, currentChunk)
		newLevelPairs = append(newLevelPairs, newKAddrPair)
	}

	return newLevelPairs
}

func InitProllyTree[K hasher.Hasher](keys []K, targetBits int) *ProllyTree[K] {
	Target := big.NewInt(1)
	Target.Lsh(Target, uint(256-targetBits))

	tree := &ProllyTree[K]{
		kvStore: make(map[string]interface{}),
		config: ProllyTreeConfig{
			Target: Target,
		},
	}

	kAddrPairs := indexValues(keys)
	sort.Sort(ByKeyHash[K](kAddrPairs))

	level := hasher.Hint(0)
	for len(kAddrPairs) > 1 {
		kAddrPairs = chunkPairs(kAddrPairs, level, tree.kvStore, tree.config)
		level += 1
	}

	tree.rootAddress = kAddrPairs[0].valueAddress

	return tree
}

func (tree ProllyTree[K]) String() string {
	var sb strings.Builder
	var queue [][32]byte
	sb.WriteString("\nTree:\n")
	queue = append(queue, tree.rootAddress)
	for len(queue) > 0 { //bfs
		nodeCount := len(queue)
		for nodeCount > 0 { //visit all nodes at the current depth
			address := queue[0]
			queue = queue[1:]

			node := tree.kvStore.Get(address).(ProllyTreeNode[K])

			for _, child := range node.children {
				keyHash := child.key.Hash()
				sb.WriteString(fmt.Sprintf("key: %v ", child.key))
				sb.WriteString(fmt.Sprintf("key hash: %s ", hex.EncodeToString(keyHash[:])))
				sb.WriteString(fmt.Sprintf("hash: %s  ", hex.EncodeToString(child.valueAddress[:])))
				queue = append(queue, child.valueAddress)
			}
			sb.WriteString("\n")
			nodeCount--
		}

		sb.WriteString("\n\n")

	}
	return sb.String()
}

func (node ProllyTreeNode[K]) Hash(seed ...byte) [32]byte {
	data := seed

	for _, kAddrPair := range node.children {
		keyHash := kAddrPair.key.Hash()

		data = append(data, keyHash[:]...)
		data = append(data, kAddrPair.valueAddress[:]...)
	}
	return sha256.Sum256(data)
}

func (t1 ProllyTree[K]) Diff(t2 ProllyTree[K]) ([]K, []K) {
	if t1.rootAddress == t2.rootAddress {
		return []K{}, []K{}
	}

	t1Root := t1.kvStore.Get(t1.rootAddress).(ProllyTreeNode[K])
	t2Root := t2.kvStore.Get(t2.rootAddress).(ProllyTreeNode[K])

	return FindNodesDiff([]ProllyTreeNode[K]{t1Root}, []ProllyTreeNode[K]{t2Root}, t1.kvStore, t2.kvStore)
}

func GetAllLeafs[K hasher.Hasher](nodes []ProllyTreeNode[K], kvStore KVStore) []ProllyTreeNode[K] {
	nodesAreLeafs := nodes[0].children[0].isLeaf()
	if nodesAreLeafs {
		return nodes
	}
	var children []ProllyTreeNode[K]

	for _, node := range nodes {
		for _, child := range node.children {
			children = append(children, child.getProllyTreeNode(kvStore))
		}
	}

	return GetAllLeafs(children, kvStore)
}

func FindNonMatchingPairs[K hasher.Hasher](t1Nodes []ProllyTreeNode[K], t2Nodes []ProllyTreeNode[K], t1KvStore KVStore, t2KvStore KVStore) (t1ExceptT2Pairs []KAddrPair[K], t2ExceptT1Pairs []KAddrPair[K]) {
	t1NodeIdx, t2NodeIdx := 0, 0
	t1NodeChildIdx, t2NodeChildIdx := 0, 0
	for t1NodeIdx < len(t1Nodes) && t2NodeIdx < len(t2Nodes) {
		t1Node := t1Nodes[t1NodeIdx]
		t2Node := t2Nodes[t2NodeIdx]
		for t1NodeChildIdx < len(t1Node.children) && t2NodeChildIdx < len(t2Node.children) {
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

	for t1NodeIdx < len(t1Nodes) {
		t1Node := t1Nodes[t1NodeIdx]
		for t1NodeChildIdx < len(t1Node.children) {
			t1NodeChild := t1Node.children[t1NodeChildIdx]
			t1ExceptT2Pairs = append(t1ExceptT2Pairs, t1NodeChild)
		}
	}
	for t2NodeIdx < len(t2Nodes) {
		t2Node := t2Nodes[t2NodeIdx]
		for t2NodeChildIdx < len(t2Node.children) {
			t2NodeChild := t2Node.children[t2NodeChildIdx]
			t2ExceptT1Pairs = append(t2ExceptT1Pairs, t2NodeChild)
		}
	}
	return
}

func FindNodesDiff[K hasher.Hasher](t1Nodes []ProllyTreeNode[K], t2Nodes []ProllyTreeNode[K], t1KvStore KVStore, t2KvStore KVStore) ([]K, []K) {
	t1NodesAreLeafs := t1Nodes[0].children[0].isLeaf()
	t2NodesAreLeafs := t2Nodes[0].children[0].isLeaf()
	if t1NodesAreLeafs && t2NodesAreLeafs {
		t1DiffKAddrPairs, t2DiffKAddrPairs := FindNonMatchingPairs(t1Nodes, t2Nodes, t1KvStore, t2KvStore)
		var t1ExceptT2pairs []K
		var t2ExceptT1Pairs []K

		for _, t1DiffKAddrPair := range t1DiffKAddrPairs {
			t1ExceptT2pairs = append(t1ExceptT2pairs, t1DiffKAddrPair.key)
		}
		for _, t2DiffKAddrPair := range t2DiffKAddrPairs {
			t2ExceptT1Pairs = append(t2ExceptT1Pairs, t2DiffKAddrPair.key)
		}

		return t1ExceptT2pairs, t2ExceptT1Pairs
	}

	var newT1Nodes []ProllyTreeNode[K]
	var newT2Nodes []ProllyTreeNode[K]
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
