package entity

import "crypto/sha256"

func StringsToBytes(args []string) [][]byte {
	res := make([][]byte, len(args))
	for i := 0; i < len(args); i++ {
		temp := []byte(args[i])
		res[i] = temp
	}
	return res
}

func ComputeRelateShardIndex(key []byte, shardAmount int) (index int) {
	hashGot := sha256.Sum256(key)
	index = (int(hashGot[0])*100 + int(hashGot[1])*10 + int(hashGot[2])) % shardAmount
	return index
}
