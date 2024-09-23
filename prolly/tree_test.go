package prolly

import (
	"crdt_sync/hasher"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func SortKVPairs(pairs []KVPair[hasher.Hstring, hasher.Hint]) {
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].key < pairs[j].key {
			return true
		}
		if pairs[i].key > pairs[j].key {
			return false
		}
		return pairs[i].value < pairs[j].value
	})
}

func AreSlicesEqual(a, b []KVPair[hasher.Hstring, hasher.Hint]) bool {
	if len(a) != len(b) {
		return false
	}
	SortKVPairs(a)
	SortKVPairs(b)
	return reflect.DeepEqual(a, b)
}

func GenerateKVPairs(nrPairs int) []KVPair[hasher.Hstring, hasher.Hint] {
	var kvPairs []KVPair[hasher.Hstring, hasher.Hint]
	for i := 0; i < nrPairs; i++ {
		kvPairs = append(kvPairs, *InitKVPair(hasher.Hstring("OLA "+strconv.Itoa(i)), hasher.Hint(i)))
	}

	return kvPairs
}

func GetShuffledSliceCopy[T any] (src []T) []T{
	a := make([]T, len(src))
	copy(a, src)
	
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}

func TestDiffSameTree(t *testing.T) {
	kvPairs := GenerateKVPairs(11)

	kvPairs1 := GetShuffledSliceCopy(kvPairs)
	tree1 := InitProllyTree(kvPairs1, 3)

	kvPairs2 := GetShuffledSliceCopy(kvPairs)
	tree2 := InitProllyTree(kvPairs2, 3)

	t1ExceptT2, t2ExceptT1 := tree1.Diff(*tree2)

	if len(t1ExceptT2) > 0 || len(t2ExceptT1) > 0 {
		t.Errorf("expected there to be no differences, found %v and %v", t1ExceptT2, t2ExceptT1)
	}

}

func TestDiffDifferentTrees(t *testing.T) {
	kvPairs := GenerateKVPairs(100000)
	kvPairs1 := GetShuffledSliceCopy(kvPairs[1000:])
	tree1 := InitProllyTree(kvPairs1, 3)

	kvPairs2 := GetShuffledSliceCopy(kvPairs[:len(kvPairs)-5000])
	tree2 := InitProllyTree(kvPairs2, 3)

	actualT1ExceptT2, actualT2ExceptT1 := tree1.Diff(*tree2)

	expectedT1ExceptT2 := kvPairs[len(kvPairs)-5000:]
	expectedT2ExceptT1 := kvPairs[:1000]

	// Check if actualT1ExceptT2 matches expectedT1ExceptT2 regardless of order
	if !AreSlicesEqual(actualT1ExceptT2, expectedT1ExceptT2) {
		t.Errorf("Expected T1 except T2 to be %v, but got %v", expectedT1ExceptT2, actualT1ExceptT2)
	}

	// Check if actualT2ExceptT1 matches expectedT2ExceptT1 regardless of order
	if !AreSlicesEqual(actualT2ExceptT1, expectedT2ExceptT1) {
		t.Errorf("Expected T2 except T1 to be %v, but got %v", expectedT2ExceptT1, actualT2ExceptT1)
	}
}
