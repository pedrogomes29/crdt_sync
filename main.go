package main

import (
	"crdt_sync/hasher"
	"crdt_sync/prolly"
	"encoding/hex"
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
	keysToInsert := GenerateKeys(100_000)
	keysToDelete := keysToInsert[:10_000]
	remainingKeys := keysToInsert[10_000:]
	treeGeneratedAtOnce := prolly.InitProllyTree(remainingKeys, 3)
	treeGeneratedIteratively := prolly.InitProllyTree([]hasher.Hstring{}, 3)
	for _, key := range keysToInsert {
		treeGeneratedIteratively.Insert(key)
	}
	fmt.Println(treeGeneratedIteratively)

	for _, key := range keysToDelete {
		keyHash := key.Hash()
		fmt.Printf("Deleting key %s with hash %s\n", key, hex.EncodeToString(keyHash[:]))
		treeGeneratedIteratively.Delete(key)
	}

	actualT1ExceptT2, actualT2ExceptT1 := treeGeneratedAtOnce.Diff(*treeGeneratedIteratively)

	if len(actualT1ExceptT2) > 0 {
		fmt.Printf("Expected T1 except T2 to be empty but got %v", actualT1ExceptT2)
	}

	if len(actualT2ExceptT1) > 0 {
		fmt.Printf("Expected T2 except T1 to be empty but got %v", actualT2ExceptT1)
	}
}
