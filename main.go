package main

import (
	"crdt_sync/hasher"
	"crdt_sync/prolly"
	"encoding/hex"
	"fmt"
	"strconv"
)

func GenerateKeys(nrKeys int) []hasher.Hstring {
	var keys []hasher.Hstring
	for i := 0; i < nrKeys; i++ {
		keys = append(keys, hasher.Hstring("OLA "+strconv.Itoa(i)))
	}

	return keys
}

func main() {
	keys := GenerateKeys(100)

	treeGeneratedAtOnce := prolly.InitProllyTree(keys, 3)

	treeGeneratedIteratively := prolly.InitProllyTree([]hasher.Hstring{}, 3)
	for _, key := range keys {
		keyHash := key.Hash()
		fmt.Printf("Inserting key %s with hash %s\n", key, hex.EncodeToString(keyHash[:]))
		treeGeneratedIteratively.Insert(key)
		fmt.Println(treeGeneratedIteratively)
	}

	fmt.Println(treeGeneratedAtOnce)
}
