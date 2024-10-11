package main

import (
	"crdt_sync/hasher"
	"crdt_sync/prolly"
	"fmt"
	"math/rand"
	"strconv"
)

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

func main() {
	keys := GenerateKeys(10000)
	randomlySortedKeys := GetShuffledSliceCopy(keys)
	treeGeneratedAtOnce := prolly.InitProllyTree(keys, 3)
	treeGeneratedIteratively := prolly.InitProllyTree([]hasher.Hstring{}, 3)
	for _, key := range randomlySortedKeys {
		treeGeneratedIteratively.Insert(key)
	}

	fmt.Print(treeGeneratedIteratively.Diff(*treeGeneratedAtOnce))
}
