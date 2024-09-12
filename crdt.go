package main

type CRDT interface {
	split() []CRDT
	join(CRDT)
	diff(CRDT) CRDT
}
