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

	actualElems := gset.Elems()

	expectedElems := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	// Compare the two sets using reflect.DeepEqual
	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}

func TestDelta(t *testing.T) {
	set1 := InitGSet[int]()
	set1.Insert(1)
	set1.Insert(2)
	set1.Insert(3)

	set2 := InitGSet[int]()
	set2.Insert(1)
	set2.Insert(2)

	setDiff := set1.Diff(*set2)

	actualElems := setDiff.Elems()

	expectedElems := map[int]struct{}{
		3: {},
	}

	if !reflect.DeepEqual(actualElems, expectedElems) {
		t.Errorf("Expected set %v, but got %v", expectedElems, actualElems)
	}
}
