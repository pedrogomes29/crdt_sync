package hasher

import (
	"testing"
)

func TestHIntEqualHash(t *testing.T) {
	int1 := Hint(1)
	int2 := Hint(1)
	if int1.Hash() != int2.Hash() {
		t.Errorf("expected hashes to be the same")
	}
}

func TestHIntDiffhash(t *testing.T) {
	int1 := Hint(1)
	int2 := Hint(2)
	if int1.Hash() == int2.Hash() {
		t.Errorf("expected hashes to be different")
	}
}

func TestHStringEqualHash(t *testing.T) {
	str1 := Hstring("hi")
	str2 := Hstring("hi")
	if str1.Hash() != str2.Hash() {
		t.Errorf("expected hashes to be the same")
	}
}

func TestHStringDiffhash(t *testing.T) {
	str1 := Hstring("hi")
	str2 := Hstring("bye")
	if str1.Hash() == str2.Hash() {
		t.Errorf("expected hashes to be different")
	}
}
