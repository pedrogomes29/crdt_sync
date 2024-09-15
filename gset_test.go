package main

import (
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	gset := InitGSet[int]()
	gset.Insert(1)
	gset.Insert(1)
	gset.Insert(2)
	gset.Insert(3)

	actualElems := gset.data

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestElems(t *testing.T) {
	gset := InitGSet[int]()
	gset.Insert(1)
	gset.Insert(1)
	gset.Insert(2)
	gset.Insert(3)

	actualElems := gset.Elems()

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestDiff(t *testing.T) {
	set1 := InitGSet[int]()
	set1.Insert(1)
	set1.Insert(2)
	set1.Insert(3)

	set2 := InitGSet[int]()
	set2.Insert(1)
	set2.Insert(2)

	setDiff := set1.Diff(*set2)

	actualElems := setDiff.data

	expectedElems := map[int]struct{}{
		3: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestIn(t *testing.T) {
	set := InitGSet[int]()
	set.Insert(1)
	set.Insert(2)
	set.Insert(3)

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	for expectedElem := range expectedElems {
		if !set.In(expectedElem) {
			t.Errorf("expected %d to be in the set", expectedElem)
		}
	}
}

func TestSplit(t *testing.T) {
	set := InitGSet[int]()
	set.Insert(1)
	set.Insert(2)
	set.Insert(3)

	actualDecomps := set.Split()

	expectedDecomp := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	if len(actualDecomps) != len(expectedDecomp) {
		t.Errorf("expected %d decompositions, got %d", len(expectedDecomp), len(actualDecomps))
	}

	for expectedElem := range expectedDecomp {
		found := false
		for _, gset := range actualDecomps {
			if _, ok := gset.data[expectedElem]; ok {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected element %d not found in actual decompositions", expectedElem)
		}
	}
}

func TestJoinSubset(t *testing.T) {
	set1 := InitGSet[int]()
	set1.Insert(1)
	set1.Insert(2)
	set1.Insert(3)

	set2 := InitGSet[int]()
	set2.Insert(1)
	set2.Insert(2)

	set1.Join(*set2)

	actualElems := set1.Elems()

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestJoinOverlapping(t *testing.T) {
	set1 := InitGSet[int]()
	set1.Insert(1)
	set1.Insert(2)
	set1.Insert(3)

	set2 := InitGSet[int]()
	set2.Insert(3)
	set2.Insert(4)

	set1.Join(*set2)

	actualElems := set1.Elems()

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestJoinDisjoint(t *testing.T) {
	set1 := InitGSet[int]()
	set1.Insert(1)
	set1.Insert(2)
	set1.Insert(3)

	set2 := InitGSet[int]()
	set2.Insert(4)
	set2.Insert(5)

	set1.Join(*set2)

	actualElems := set1.Elems()

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}
