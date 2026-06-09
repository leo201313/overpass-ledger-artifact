package stateManager

import (
	"bytes"
	"opl/common"
	"testing"
)

func TestVCT(t *testing.T) {
	state1 := common.BytesToAddress([]byte("state1"))
	state2 := common.BytesToAddress([]byte("state2"))
	state3 := common.BytesToAddress([]byte("state3"))

	version1 := common.BytesToHash([]byte("version1"))
	version2 := common.BytesToHash([]byte("version2"))
	version3 := common.BytesToHash([]byte("version3"))
	version4 := common.BytesToHash([]byte("version4"))

	val1 := []byte("value1")
	val2 := []byte("value2")
	val3 := []byte("value3")
	val4 := []byte("value4")

	vct := NewVCT(common.BytesToHash([]byte("test vct")))

	// test for write and read
	vct.WriteState(state1, val1, version1, 1)
	vct.WriteState(state2, val2, version2, 1)
	vct.WriteState(state3, val3, version3, 1)

	have, readValue, readVersion, _ := vct.ReadState(state1)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val1) {
		t.Fatal()
	}
	if readVersion != version1 {
		t.Fatal()
	}

	have, readValue, readVersion, _ = vct.ReadState(state2)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val2) {
		t.Fatal()
	}
	if readVersion != version2 {
		t.Fatal()
	}

	have, readValue, readVersion, _ = vct.ReadState(state3)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val3) {
		t.Fatal()
	}
	if readVersion != version3 {
		t.Fatal()
	}

	// test for update
	vct.WriteState(state1, val4, version4, 1)
	have, readValue, readVersion, _ = vct.ReadState(state1)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val4) {
		t.Fatal()
	}
	if readVersion != version4 {
		t.Fatal()
	}

}
