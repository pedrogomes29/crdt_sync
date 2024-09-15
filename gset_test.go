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

func TestDiff(t *testing.T) {
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

func TestSplit(t *testing.T) {
	set := InitGSet[int]()
	set.Insert(1)
	set.Insert(2)
	set.Insert(3)

	actualDecomps := set.Split()

	var expectedDecomps []GSet[int]
	expectedDecomps = append(expectedDecomps, *InitGSetWithIrrElem(1))
	expectedDecomps = append(expectedDecomps, *InitGSetWithIrrElem(2))
	expectedDecomps = append(expectedDecomps, *InitGSetWithIrrElem(3))

	if !reflect.DeepEqual(actualDecomps, expectedDecomps) {
		t.Errorf("Expected set %v, but got %v", actualDecomps, expectedDecomps)
	}

}
