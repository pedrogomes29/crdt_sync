package prolly

import (
	"bytes"
	"crdt_sync/hasher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"slices"
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

func (p KAddrPair[K]) getProllyTreeNode(kvStore KVStore[ProllyTreeNode[K]]) ProllyTreeNode[K] {
	return kvStore.Get(p.valueAddress)
}

type ProllyTreeNode[K hasher.Hasher] struct {
	children []KAddrPair[K]
}

type ProllyTreeConfig struct {
	Target *big.Int
}

type ProllyTree[K hasher.Hasher] struct {
	rootAddress [32]byte
	kvStore     KVStore[ProllyTreeNode[K]]
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

func chunkPairs[K hasher.Hasher](pairs []KAddrPair[K], level hasher.Hint, kvStore KVStore[ProllyTreeNode[K]], config ProllyTreeConfig) []KAddrPair[K] {
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
		kvStore: make(map[string]ProllyTreeNode[K]),
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

	if len(kAddrPairs) == 1 {
		tree.rootAddress = kAddrPairs[0].valueAddress
	}

	return tree
}

func (tree *ProllyTree[K]) IsEmpty() bool {
	return reflect.ValueOf(tree.rootAddress).IsZero()
}

func (tree *ProllyTree[K]) Insert(key K) {
	var newRoot *ProllyTreeNode[K]
	if tree.IsEmpty() {
		newRoot = &ProllyTreeNode[K]{
			[]KAddrPair[K]{
				{
					key: key,
				},
			},
		}
	} else {
		oldRoot := tree.kvStore.Get(tree.rootAddress)
		keyHash := key.Hash()
		oldRootBiggestKey := oldRoot.children[len(oldRoot.children)-1].key
		oldRootBiggestKeyHash := oldRootBiggestKey.Hash()
		newLargestElement := bytes.Compare(keyHash[:], oldRootBiggestKeyHash[:]) == 1

		newBoundaryNode, level := oldRoot.insert(key, tree.kvStore, tree.config)
		for newBoundaryNode != nil {
			newRoot = &ProllyTreeNode[K]{
				[]KAddrPair[K]{
					{
						oldRoot.children[len(oldRoot.children)-1].key,
						oldRoot.Hash(),
					},
				},
			}
			newKAddrPair := KAddrPair[K]{
				key,
				newBoundaryNode.Hash(),
			}

			var posToInsert int
			if newLargestElement {
				posToInsert = 1
			} else {
				posToInsert = 0
			}
			newBoundaryNode = newRoot.insertKAddrPair(posToInsert, newKAddrPair, hasher.Hint(level+1), tree.config)
			if newBoundaryNode != nil {
				tree.kvStore.Put(newRoot.Hash(), *newRoot)
				tree.kvStore.Put(newBoundaryNode.Hash(), *newBoundaryNode)
				level++
			}
			oldRoot = *newRoot
		}
		newRoot = &oldRoot
	}
	tree.kvStore.Put(newRoot.Hash(), *newRoot)
	tree.rootAddress = newRoot.Hash()
}

func (node *ProllyTreeNode[K]) insert(key K, kvStore KVStore[ProllyTreeNode[K]], config ProllyTreeConfig) (newBoundaryNode *ProllyTreeNode[K], nodeLevel int) {
	keyHash := key.Hash()
	firstBiggerKeyIdx := len(node.children)

	for childIdx, kAddrPair := range node.children {
		childHash := kAddrPair.key.Hash()
		comparison := bytes.Compare(keyHash[:], childHash[:])
		if comparison == 0 {
			return nil, -1
		}
		if comparison < 0 { //if new key is smaller than the current key
			firstBiggerKeyIdx = childIdx
			break
		}
	}
	outdatedNodeHash := node.Hash()

	if node.isLeaf() {
		kAddrPair := KAddrPair[K]{
			key: key,
		}
		nodeLevel = 0
		newBoundaryNode = node.insertKAddrPair(firstBiggerKeyIdx, kAddrPair, hasher.Hint(nodeLevel), config)
	} else {
		kAddrPair := node.children[min(firstBiggerKeyIdx, len(node.children)-1)]
		child := kvStore.Get(kAddrPair.valueAddress)
		newChildBoundaryNode, childLevel := child.insert(key, kvStore, config)
		nodeLevel = childLevel + 1
		var oldChildIdx int
		if newChildBoundaryNode != nil {
			kAddrPair := KAddrPair[K]{
				key:          key,
				valueAddress: newChildBoundaryNode.Hash(),
			}
			isInsertAtEnd := firstBiggerKeyIdx == len(node.children)
			//new child is at the left when inserting in the middle and at the right when inserting at the end
			if isInsertAtEnd {
				oldChildIdx = firstBiggerKeyIdx - 1
			} else {
				oldChildIdx = firstBiggerKeyIdx + 1
			}
			newBoundaryNode = node.insertKAddrPair(firstBiggerKeyIdx, kAddrPair, hasher.Hint(nodeLevel), config)

			if newBoundaryNode != nil && !isInsertAtEnd {
				//if there was a split when inserting at the current node and the new child is at the left,
				//compensate the old child idx by removing the number of keys that are now part of the new node
				oldChildIdx -= len(newBoundaryNode.children)
			}

		} else {
			oldChildIdx = firstBiggerKeyIdx
			if firstBiggerKeyIdx == len(node.children) {
				oldChildIdx--
			}
		}
		node.children[oldChildIdx].key = child.children[len(child.children)-1].key
		node.children[oldChildIdx].valueAddress = child.Hash()
	}

	if newBoundaryNode != nil {
		kvStore.Put(newBoundaryNode.Hash(), *newBoundaryNode)
	}

	kvStore.Delete(outdatedNodeHash)
	kvStore.Put(node.Hash(), *node)

	return
}

func (node *ProllyTreeNode[K]) insertKAddrPair(pairIdx int, newKAddrPair KAddrPair[K], level hasher.Hint, config ProllyTreeConfig) *ProllyTreeNode[K] {
	node.children = slices.Insert(node.children, pairIdx, newKAddrPair)
	var newNode *ProllyTreeNode[K]

	if pairIdx == len(node.children)-1 { //if key is inserted at the end
		lastKAddrPairB4Insert := node.children[len(node.children)-2]
		if lastKAddrPairB4Insert.CheckBoundary(level, config.Target) {
			newNode = &ProllyTreeNode[K]{}

			newNode.children = make([]KAddrPair[K], pairIdx)
			copy(newNode.children, node.children[pairIdx:])

			node.children = node.children[:pairIdx]
		}
	} else if newKAddrPair.CheckBoundary(level, config.Target) {
		newNode = &ProllyTreeNode[K]{}

		newNode.children = make([]KAddrPair[K], len(node.children[:pairIdx+1]))
		copy(newNode.children, node.children[:pairIdx+1])

		node.children = node.children[pairIdx+1:]
	}
	return newNode
}

func (node ProllyTreeNode[K]) String(kvStore KVStore[ProllyTreeNode[K]]) string {
	var sb strings.Builder
	var queue [][32]byte
	level := 0
	queue = append(queue, node.Hash())
	for len(queue) > 0 { //bfs
		nodeCount := len(queue)
		sb.WriteString(fmt.Sprintf("Level %d: \n", level))
		for nodeCount > 0 { //visit all nodes at the current depth
			address := queue[0]
			queue = queue[1:]
			node := kvStore.Get(address)

			for _, child := range node.children {
				keyHash := child.key.Hash()
				sb.WriteString(fmt.Sprintf("key: %v, ", child.key))
				sb.WriteString(fmt.Sprintf("key hash: %s ", hex.EncodeToString(keyHash[:])))
				if !child.isLeaf() {
					sb.WriteString(fmt.Sprintf("Key: %v, Address: %s\n", child.key, hex.EncodeToString(child.valueAddress[:])))
					queue = append(queue, child.valueAddress)
				}
			}
			sb.WriteString("\n")
			nodeCount--
		}

		sb.WriteString("\n\n")
		level++
	}
	return sb.String()
}

func (tree ProllyTree[K]) String() string {
	var sb strings.Builder
	if tree.IsEmpty() {
		return "tree is empty"
	}

	sb.WriteString("\nTree:\n")

	root := tree.kvStore.Get(tree.rootAddress)
	return root.String(tree.kvStore)
}

func (node ProllyTreeNode[K]) isLeaf() bool {
	return node.children[0].isLeaf()
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

	t1Root := t1.kvStore.Get(t1.rootAddress)
	t2Root := t2.kvStore.Get(t2.rootAddress)

	return FindNodesDiff([]ProllyTreeNode[K]{t1Root}, []ProllyTreeNode[K]{t2Root}, t1.kvStore, t2.kvStore)
}

func GetAllLeafs[K hasher.Hasher](nodes []ProllyTreeNode[K], kvStore KVStore[ProllyTreeNode[K]]) []ProllyTreeNode[K] {
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

func FindNonMatchingPairs[K hasher.Hasher](t1Nodes []ProllyTreeNode[K], t2Nodes []ProllyTreeNode[K], t1KvStore KVStore[ProllyTreeNode[K]], t2KvStore KVStore[ProllyTreeNode[K]]) (t1ExceptT2Pairs []KAddrPair[K], t2ExceptT1Pairs []KAddrPair[K]) {
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

func FindNodesDiff[K hasher.Hasher](t1Nodes []ProllyTreeNode[K], t2Nodes []ProllyTreeNode[K], t1KvStore KVStore[ProllyTreeNode[K]], t2KvStore KVStore[ProllyTreeNode[K]]) ([]K, []K) {
	t1NodesAreLeafs := t1Nodes[0].isLeaf()
	t2NodesAreLeafs := t2Nodes[0].isLeaf()
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
		newT2Nodes = GetAllLeafs(t2Nodes, t2KvStore)
	}

	return FindNodesDiff(newT1Nodes, newT2Nodes, t1KvStore, t2KvStore)
}
