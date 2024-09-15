package main

import (
	"bytes"
	"testing"
)

func TestHIntEqualHash(t *testing.T) {
	int1 := hint(1)
	int2 := hint(1)
	if !bytes.Equal(int1.Hash(), int2.Hash()) {
		t.Errorf("expected hashes to be the same")
	}
}

func TestHIntDiffhash(t *testing.T) {
	int1 := hint(1)
	int2 := hint(2)
	if bytes.Equal(int1.Hash(), int2.Hash()) {
		t.Errorf("expected hashes to be different")
	}
}

func TestHStringEqualHash(t *testing.T) {
	str1 := hstring("hi")
	str2 := hstring("hi")
	if !bytes.Equal(str1.Hash(), str2.Hash()) {
		t.Errorf("expected hashes to be the same")
	}
}

func TestHStringDiffhash(t *testing.T) {
	str1 := hstring("hi")
	str2 := hstring("bye")
	if bytes.Equal(str1.Hash(), str2.Hash()) {
		t.Errorf("expected hashes to be different")
	}
}
