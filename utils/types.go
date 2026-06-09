package utils

import (
	"encoding/binary"
	"fmt"
)

// Uint64ToBytes converts an uint64 num into a bytes array
func Uint64ToBytes(num uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(num))
	return b
}

// BytesToUint64 converts a bytes array into uint64 num
func BytesToUint64(bytes []byte) (uint64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("fail in BytesToUint64: the bytes array length should be 8 but get %d", len(bytes))
	}
	i := binary.LittleEndian.Uint64(bytes)
	return i, nil
}

// XorBytes returns the results of two byte arrays after xor operation
func XorBytes(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("fail in utils.XorBytes: got byte array with unmatch length: %d and %d", len(a), len(b))
	}
	res := make([]byte, len(b))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] ^ b[i]
	}
	return res, nil
}
