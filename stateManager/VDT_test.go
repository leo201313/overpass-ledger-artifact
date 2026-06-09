package stateManager

import (
	"bytes"
	"opl/common"
	"opl/elements"
	"testing"
)

func TestVDT(t *testing.T) {
	vdt := NewVDT()

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
	val5 := []byte("value5")
	val6 := []byte("value6")
	val7 := []byte("value7")
	//val8 := []byte("value8")
	//val9 := []byte("value9")

	vdt.WriteState(state1, val1, version1)
	vdt.WriteState(state2, val2, version1)
	vdt.WriteState(state3, val3, version1)

	// test for read
	have, readValue, readVersion := vdt.ReadState(state1)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val1) {
		t.Fatal()
	}
	if readVersion != version1 {
		t.Fatal()
	}

	// test for update
	vdt.WriteState(state1, val4, version2)
	have, readValue, readVersion = vdt.ReadState(state1)
	if !have {
		t.Fatal()
	}
	if !bytes.Equal(readValue, val4) {
		t.Fatal()
	}
	if readVersion != version2 {
		t.Fatal()
	}

	// test for trim

	committedStates := []elements.StateCommit{
		{
			Version: version2,
			Address: state2,
			Value:   nil,
		},
	}

	vdt.TrimState(committedStates, common.Hash{})
	have, readValue, readVersion = vdt.ReadState(state2)
	if have {
		t.Fatal()
	}

	vdt.WriteState(state2, val5, version3)
	vdt.WriteState(state2, val6, version3)
	vdt.WriteState(state2, val7, version3)

	committedStates = []elements.StateCommit{
		{
			Version: version4,
			Address: state3,
			Value:   nil,
		},
	}

	vdt.TrimState(committedStates, common.Hash{})
	have, readValue, readVersion = vdt.ReadState(state2)
	if !have {
		t.Fatal()
	}

	if !bytes.Equal(readValue, val7) {
		t.Fatal()
	}
	if readVersion != version3 {
		t.Fatal()
	}

	have, readValue, readVersion = vdt.ReadState(state1)
	if have {
		t.Fatal()
	}

}
