package prolly

import (
	"crdt_sync/hasher"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func SortByKeys(pairs []hasher.Hstring) {
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i] < pairs[j] {
			return true
		}
		if pairs[i] > pairs[j] {
			return false
		}
		return pairs[i] < pairs[j]
	})
}

func AreSlicesEqual(a, b []hasher.Hstring) bool {
	if len(a) != len(b) {
		return false
	}
	SortByKeys(a)
	SortByKeys(b)
	return reflect.DeepEqual(a, b)
}

func GenerateKeys(nrKeys int) []hasher.Hstring {
	var keys []hasher.Hstring
	for i := 0; i < nrKeys; i++ {
		keys = append(keys, hasher.Hstring("OLA "+strconv.Itoa(i)))
	}

	return keys
}

func GetShuffledSliceCopy[T any](src []T) []T {
	a := make([]T, len(src))
	copy(a, src)

	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}

func TestDiffSameTree(t *testing.T) {
	keys := GenerateKeys(11)

	keys1 := GetShuffledSliceCopy(keys)
	tree1 := InitProllyTree(keys1, 3)

	keys2 := GetShuffledSliceCopy(keys)
	tree2 := InitProllyTree(keys2, 3)

	t1ExceptT2, t2ExceptT1 := tree1.Diff(*tree2)

	if len(t1ExceptT2) > 0 || len(t2ExceptT1) > 0 {
		t.Errorf("expected there to be no differences, found %v and %v", t1ExceptT2, t2ExceptT1)
	}

}

func TestDiffDifferentTrees(t *testing.T) {
	keys := GenerateKeys(100000)

	nrKeysT1ExceptT2 := 5
	nrKeysT2ExceptT1 := 5

	keys1 := GetShuffledSliceCopy(keys[nrKeysT2ExceptT1:])
	tree1 := InitProllyTree(keys1, 3)

	keys2 := GetShuffledSliceCopy(keys[:len(keys)-nrKeysT1ExceptT2])
	tree2 := InitProllyTree(keys2, 3)

	actualT1ExceptT2, actualT2ExceptT1 := tree1.Diff(*tree2)

	expectedT1ExceptT2 := keys[len(keys)-nrKeysT1ExceptT2:]
	expectedT2ExceptT1 := keys[:nrKeysT2ExceptT1]

	// Check if actualT1ExceptT2 matches expectedT1ExceptT2 regardless of order
	if !AreSlicesEqual(actualT1ExceptT2, expectedT1ExceptT2) {
		t.Errorf("Expected T1 except T2 to be %v, but got %v", expectedT1ExceptT2, actualT1ExceptT2)
	}

	// Check if actualT2ExceptT1 matches expectedT2ExceptT1 regardless of order
	if !AreSlicesEqual(actualT2ExceptT1, expectedT2ExceptT1) {
		t.Errorf("Expected T2 except T1 to be %v, but got %v", expectedT2ExceptT1, actualT2ExceptT1)
	}
}
