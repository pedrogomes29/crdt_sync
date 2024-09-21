package main

import (
	"crdt_sync/hasher"
	"crdt_sync/prolly"
	"fmt"
	"strconv"
)

func main() {
	var kvPairs []prolly.KVPair[hasher.Hstring, hasher.Hint]
	for i := 0; i < 10; i++ {
		kvPairs = append(kvPairs, *prolly.InitKVPair(hasher.Hstring("OLA "+strconv.Itoa(i)), hasher.Hint(i)))
	}

	tree := prolly.InitProllyTree(kvPairs)
	fmt.Printf("%s", tree.String())
}
