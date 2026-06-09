package utils

import (
	"bytes"
	"testing"
)

func TestUint64ToBytes(t *testing.T) {
	n1 := uint64(10000000019090)
	n2 := uint64(123456789)
	n3 := uint64(7813333)

	n1_t, _ := BytesToUint64(Uint64ToBytes(n1))
	n2_t, _ := BytesToUint64(Uint64ToBytes(n2))
	n3_t, _ := BytesToUint64(Uint64ToBytes(n3))

	if n1_t != n1 || n2_t != n2 || n3_t != n3 {
		t.Fail()
	}
}

func TestXorBytes(t *testing.T) {
	a := []byte{0x00, 0x10, 0x00}
	b := []byte{0x10, 0x10, 0x00}

	expect := []byte{0x10, 0x00, 0x00}

	res, err := XorBytes(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, expect) {
		t.Fail()
	}
}
