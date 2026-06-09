package stateManager

import (
	"bytes"
	"testing"
)

func TestEpochToBytes(t *testing.T) {
	key1 := EpochToBytes(23)
	key2 := EpochToBytes(24)
	key3 := EpochToBytes(25)
	key4 := EpochToBytes(23)
	if bytes.Equal(key1, key2) {
		t.Fatal()
	}
	if bytes.Equal(key2, key3) {
		t.Fatal()
	}
	if bytes.Equal(key3, key1) {
		t.Fatal()
	}
	if !bytes.Equal(key1, key4) {
		t.Fatal()
	}
}

func TestBytesToUint64(t *testing.T) {
	var num1 uint64 = 8
	var num2 uint64 = 9
	var num3 uint64 = 10

	num1_, err := BytesToUint64(Uint64ToBytes(num1))
	if err != nil {
		t.Fatal()
	}

	num2_, err := BytesToUint64(Uint64ToBytes(num2))
	if err != nil {
		t.Fatal()
	}

	num3_, err := BytesToUint64(Uint64ToBytes(num3))
	if err != nil {
		t.Fatal()
	}

	if num1_ != num1 {
		t.Fatal()
	}
	if num2_ != num2 {
		t.Fatal()
	}
	if num3_ != num3 {
		t.Fatal()
	}

}
