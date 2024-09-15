package main

type CRDT interface {
	split() []CRDTDecomposition
	join(CRDT)
	diff(CRDT) CRDT
}

type CRDTDecomposition interface {
	CRDT
	Hasher
}
