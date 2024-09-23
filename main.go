package main

import (
	"crdt_sync/hasher"
	"crdt_sync/prolly"
	"fmt"
)

func main() {
	var kvPairs1 []prolly.KVPair[hasher.Hstring, hasher.Hint]
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 10"), hasher.Hint(10)))
	kvPairs1 = append(kvPairs1, *prolly.InitKVPair(hasher.Hstring("OLA 11"), hasher.Hint(11)))

	tree1 := prolly.InitProllyTree(kvPairs1)

	var kvPairs2 []prolly.KVPair[hasher.Hstring, hasher.Hint]
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 4"), hasher.Hint(4)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 2"), hasher.Hint(2)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 3"), hasher.Hint(3)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 1"), hasher.Hint(1)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 9"), hasher.Hint(9)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 7"), hasher.Hint(7)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 8"), hasher.Hint(8)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 5"), hasher.Hint(5)))
	kvPairs2 = append(kvPairs2, *prolly.InitKVPair(hasher.Hstring("OLA 6"), hasher.Hint(6)))

	tree2 := prolly.InitProllyTree(kvPairs2)

	t1ExceptT2, t2ExceptT1 := tree1.Diff(*tree2)

	fmt.Printf("%v %v\n", t1ExceptT2, t2ExceptT1)
}
