package stateManager

import (
	"bytes"
	"opl/coes"
	"opl/rlp"
)

func EpochToBytes(height uint64) []byte {
	epochPrefix, _ := rlp.EncodeToBytes([]byte(coes.EpochPrefix))
	heightBytes := Uint64ToBytes(height)
	tot := bytes.Join([][]byte{epochPrefix, heightBytes}, []byte{})
	return tot
}

func Uint64ToBytes(num uint64) []byte {
	uint64Bytes, _ := rlp.EncodeToBytes(num)
	return uint64Bytes
}

func BytesToUint64(bts []byte) (uint64, error) {
	var num uint64
	err := rlp.DecodeBytes(bts, &num)
	return num, err
}
