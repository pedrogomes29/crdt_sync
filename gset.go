package main

type GSet[T comparable] struct {
	data map[T]struct{}
}

type GSetDecomposition[T comparable] struct {
	GSet[T]
}

func InitGSetDecomp[T comparable](irrElem T) *GSetDecomposition[T] {
	gset := &GSetDecomposition[T]{
		GSet: GSet[T]{
			data: make(map[T]struct{}),
		},
	}
	gset.data[irrElem] = struct{}{}
	return gset
}

func InitGSet[T comparable]() *GSet[T] {
	return &GSet[T]{
		data: make(map[T]struct{}),
	}
}

func (set *GSet[T]) In(elem T) bool {
	_, elemInSet := set.data[elem]
	return elemInSet
}

func (set *GSet[T]) Elems() map[T]struct{} {
	return set.data
}

func (set *GSet[T]) Insert(elem T) GSet[T] {
	oldSet := *set
	set.data[elem] = struct{}{}
	return set.Diff(oldSet)
}

func (set *GSet[T]) Split() []GSetDecomposition[T] {
	//Go doesn't allow custom types to implement comparable, so set[GSet[T]] isn't allowed
	//because of this, decompositions are returned in a random order
	var joinDecompositions []GSetDecomposition[T]
	for elem := range set.data {
		joinDecomposition := InitGSetDecomp[T](elem)
		joinDecompositions = append(joinDecompositions, *joinDecomposition)
	}
	return joinDecompositions
}

func (set *GSet[T]) Join(delta GSet[T]) {
	for elem := range delta.data {
		set.data[elem] = struct{}{}
	}
}

func (set *GSet[T]) Diff(delta GSet[T]) GSet[T] {
	joinDecompositions := set.Split()
	diff := InitGSet[T]()
	for _, decomposition := range joinDecompositions {
		for elem := range decomposition.data {
			if _, elemInSet := delta.data[elem]; !elemInSet {
				diff.data[elem] = struct{}{}
			}
		}
	}
	return *diff
}
