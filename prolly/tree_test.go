package prolly

import (
	"crdt_sync/hasher"
	"reflect"
	"sort"
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

func TestDiffSameTree(t *testing.T) {
	var kvPairs1 []KVPair[hasher.Hstring, hasher.Hint]
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 1"), hasher.Hint(1)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 2"), hasher.Hint(2)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 10"), hasher.Hint(10)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 11"), hasher.Hint(11)))

	tree1 := InitProllyTree(kvPairs1)

	var kvPairs2 []KVPair[hasher.Hstring, hasher.Hint]
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 2"), hasher.Hint(2)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 1"), hasher.Hint(1)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 11"), hasher.Hint(11)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 10"), hasher.Hint(10)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))

	tree2 := InitProllyTree(kvPairs2)

	t1ExceptT2, t2ExceptT1 := tree1.Diff(*tree2)

	if len(t1ExceptT2) > 0 || len(t2ExceptT1) > 0 {
		t.Errorf("expected there to be no differences, found %v and %v", t1ExceptT2, t2ExceptT1)
	}

}

func TestDiffDifferentTrees(t *testing.T) {
	var kvPairs1 []KVPair[hasher.Hstring, hasher.Hint]
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 10"), hasher.Hint(10)))
	kvPairs1 = append(kvPairs1, *InitKVPair(hasher.Hstring("OLA 11"), hasher.Hint(11)))

	tree1 := InitProllyTree(kvPairs1)

	var kvPairs2 []KVPair[hasher.Hstring, hasher.Hint]
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 2"), hasher.Hint(2)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 1"), hasher.Hint(1)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs2 = append(kvPairs2, *InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))

	tree2 := InitProllyTree(kvPairs2)

	actualT1ExceptT2, actualT2ExceptT1 := tree1.Diff(*tree2)
	expectedT1ExceptT2 := []KVPair[hasher.Hstring, hasher.Hint]{
		*InitKVPair(hasher.Hstring("OLA 10"), hasher.Hint(10)),
		*InitKVPair(hasher.Hstring("OLA 11"), hasher.Hint(11)),
	}
	expectedT2ExceptT1 := []KVPair[hasher.Hstring, hasher.Hint]{
		*InitKVPair(hasher.Hstring("OLA 1"), hasher.Hint(1)),
		*InitKVPair(hasher.Hstring("OLA 2"), hasher.Hint(2)),
	}

	// Check if actualT1ExceptT2 matches expectedT1ExceptT2 regardless of order
	if !AreSlicesEqual(actualT1ExceptT2, expectedT1ExceptT2) {
		t.Errorf("Expected T1 except T2 to be %v, but got %v", expectedT1ExceptT2, actualT1ExceptT2)
	}

	// Check if actualT2ExceptT1 matches expectedT2ExceptT1 regardless of order
	if !AreSlicesEqual(actualT2ExceptT1, expectedT2ExceptT1) {
		t.Errorf("Expected T2 except T1 to be %v, but got %v", expectedT2ExceptT1, actualT2ExceptT1)
	}
}
