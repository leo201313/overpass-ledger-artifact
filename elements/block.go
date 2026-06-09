package elements

import (
	"bytes"
	"crypto/sha256"
	"opl/common"
	"opl/rlp"
)

type Block struct {
	BlockID common.Hash

	ShardNO uint8       // the number of the shard where this block is uploaded
	Version common.Hash // the version of which this block is based on in the shard
	Nonce   uint64

	Transactions []Transaction
}

func (b *Block) Hash() common.Hash {
	shardNO, _ := rlp.EncodeToBytes(b.ShardNO)
	version, _ := rlp.EncodeToBytes(b.Version)
	nonce, _ := rlp.EncodeToBytes(b.Nonce)

	jointBytes := bytes.Join([][]byte{shardNO, version, nonce}, []byte{})
	hashBytes := sha256.Sum256(jointBytes)
	return hashBytes
}

func (b *Block) SetBlockID() {
	b.BlockID = b.Hash()
}
