package sync

import "crdt_sync/hasher"

type CRDT interface {
	split() []CRDTDecomposition
	join(CRDT)
	diff(CRDT) CRDT
}

type CRDTDecomposition interface {
	CRDT
	hasher.Hasher
}
