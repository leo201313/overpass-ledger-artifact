package elements

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"opl/common"
	"opl/rlp"
	"strings"
)

type Epoch struct {
	EpochID common.Hash

	PreviousEpoch           common.Hash
	Height                  uint64
	BlockSequence           []common.Hash
	BlockAssociatedShardNOs []uint8

	Receipts       []BlockReceipt
	StateCommitSet []StateCommit
}

type BlockReceipt struct {
	BlockID   common.Hash
	TXIDs     []common.Hash
	TXProcess []uint8 // 0 is inherited, 1 is re-executed
	TXResults [][]byte
}

type StateCommit struct {
	Version common.Hash
	Address common.Address
	Value   []byte
}

func (ep *Epoch) Hash() common.Hash {
	previous, _ := rlp.EncodeToBytes(ep.PreviousEpoch)
	height, _ := rlp.EncodeToBytes(ep.Height)
	blockSequence, _ := rlp.EncodeToBytes(ep.BlockSequence)
	shardNOs, _ := rlp.EncodeToBytes(ep.BlockAssociatedShardNOs)
	jointBytes := bytes.Join([][]byte{previous, height, blockSequence, shardNOs}, []byte{})
	hashBytes := sha256.Sum256(jointBytes)
	return hashBytes
}

func (ep *Epoch) SetEpochID() {
	ep.EpochID = ep.Hash()
}

func (ep *Epoch) BackAllTxIDs() []common.Hash {
	txIDs := make([]common.Hash, 0)
	for _, receipt := range ep.Receipts {
		txIDs = append(txIDs, receipt.TXIDs...)
	}
	return txIDs
}

func (ep *Epoch) BackAllTxIDsAndHandleType() (txIDs []common.Hash, handleTypes []uint8) {
	for _, receipt := range ep.Receipts {
		txIDs = append(txIDs, receipt.TXIDs...)
		handleTypes = append(handleTypes, receipt.TXProcess...)
	}
	return
}

func (ep *Epoch) String() string {
	txCount := 0
	for i := 0; i < len(ep.Receipts); i++ {
		txCount += len(ep.Receipts[i].TXResults)
	}
	return fmt.Sprintf(
		"Epoch Details:\n"+
			"  EpochID: %x\n"+
			"  PreviousEpoch: %x\n"+
			"  Height: %d\n"+
			"  BlockSequence: [%s]\n"+
			"  BlockAssociatedShardNOs: [%s]\n"+
			"  Receipts Count: %d\n"+
			"  Transactions Count: %d\n"+
			"  StateCommitSet Count: %d\n",
		ep.EpochID,
		ep.PreviousEpoch,
		ep.Height,
		formatHashes(ep.BlockSequence),
		formatShardNOs(ep.BlockAssociatedShardNOs),
		len(ep.Receipts),
		txCount,
		len(ep.StateCommitSet),
	)
}

func formatHashes(hashes []common.Hash) string {
	var hashStrings []string
	for _, ha := range hashes {
		hashStrings = append(hashStrings, fmt.Sprintf("%x", ha))
	}
	return strings.Join(hashStrings, ", ")
}

func formatShardNOs(shardNOs []uint8) string {
	var shardStrings []string
	for _, shard := range shardNOs {
		shardStrings = append(shardStrings, fmt.Sprintf("%d", shard))
	}
	return strings.Join(shardStrings, ", ")
}
