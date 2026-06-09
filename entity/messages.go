package entity

import (
	"fmt"
	"opl/common"
	"opl/elements"
	"strings"
)

type TxMsg struct {
	Content elements.Transaction
}

// TxGroupMsg used for transmit huge amount of transactions
type TxGroupMsg struct {
	Content []elements.Transaction
}

type LowPreprepareMsg struct {
	ShardNO    uint8
	EpochID    common.Hash
	Height     uint64
	Nonce      uint64
	TxSequence []common.Hash
}

type LowPrepareMsg struct {
	Height uint64
	Nonce  uint64

	// TODO: For test, the following is not implemented
	//ShardNO uint8
	//EpochID common.Hash
	//From      common.Address
	//Signature []byte
}

type LowCommitMsg struct {
	Height uint64
	Nonce  uint64

	// TODO: For test, the following is not implemented
	//ShardNO uint8
	//EpochID common.Hash
	//From      common.Address
	//Signature []byte
}

type LowBlockMsg struct {
	Block elements.Block
}

type LowDoneMsg struct {
	Height uint64
	Nonce  uint64

	// TODO: For test, the following is not implemented
	//ShardNO uint8
	//EpochID common.Hash
	//From      common.Address
	//Signature []byte
}

type LowEpochStepInMsg struct {
	PrevHeight uint64
	NowHeight  uint64
	EpochIDs   []common.Hash

	// TODO: For test, the following is not implemented
	//ShardNO uint8
	//From      common.Address
	//CurrentTurn int
	//Signature []byte
}

type LowEpochDoneMsg struct {
	PrevHeight uint64
	NowHeight  uint64

	// TODO: For test, the following is not implemented
	//ShardNO uint8
	//From      common.Address
	//CurrentTurn int
	//Signature []byte
}

type UppPreprepareMsg struct {
	EpochID         common.Hash
	PreviousVersion common.Hash
	Height          uint64
	BlockSequence   []common.Hash
	BlockAssociated []uint8

	// TODO: For test, the following is not implemented
	//From      common.Address
	//Signature []byte
}

type UppPrepareMsg struct {
	EpochID common.Hash

	// TODO: For test, the following is not implemented
	//From      common.Address
	//Signature []byte
}

type UppCommitMsg struct {
	EpochID common.Hash

	// TODO: For test, the following is not implemented
	// From      common.Address
	// Signature []byte

}

type UppDoneMsg struct {
	EpochID common.Hash

	// TODO: For test, the following is not implemented
	//Receipt common.Hash
	//From      common.Address
	//Signature []byte
}

type UppEpochMsg struct {
	Epoch elements.Epoch
}

//type UppTickMsg struct {
//	BaseVersion common.Hash
//	NextVersion common.Hash
//}

type LowMulticastUppTickMsg struct {
	BaseVersion common.Hash
	NextVersion common.Hash
}

type TestEpochCommitMsg struct {
	Epoch elements.Epoch
}

type TestBlocksUploadMsg struct {
	Blocks []elements.Block
}

type TestStateMsg struct {
	State    uint8
	TimeUsed uint64
	TxNum    uint64
	BlockNum uint64
}

// String method for LowPreprepareMsg
// Returns a formatted string representation of the LowPreprepareMsg struct.
func (l LowPreprepareMsg) String() string {
	var builder strings.Builder

	// Add information about the shard number
	builder.WriteString(fmt.Sprintf("ShardNO: %d\n", l.ShardNO))

	// Add epoch ID (assuming common.Hash has a valid String method)
	builder.WriteString(fmt.Sprintf("EpochID: %x\n", l.EpochID))

	// Add height and nonce information
	builder.WriteString(fmt.Sprintf("Height: %d\n", l.Height))
	builder.WriteString(fmt.Sprintf("Nonce: %d\n", l.Nonce))

	// Add transaction sequence (if any)
	builder.WriteString(fmt.Sprintf("TxSequence: %d\n", len(l.TxSequence)))
	return builder.String()
}

// String method for LowEpochStepInMsg
// Returns a formatted string representation of the LowEpochStepInMsg struct.
func (l LowEpochStepInMsg) String() string {
	var builder strings.Builder

	// Add previous and current height information
	builder.WriteString(fmt.Sprintf("PrevHeight: %d\n", l.PrevHeight))
	builder.WriteString(fmt.Sprintf("NowHeight: %d\n", l.NowHeight))

	// Add EpochIDs list (if any)
	builder.WriteString("EpochIDs: ")
	if len(l.EpochIDs) > 0 {
		for _, epochID := range l.EpochIDs {
			builder.WriteString(fmt.Sprintf("%x ", epochID)) // Assuming common.Hash has a valid String method
		}
	} else {
		builder.WriteString("No EpochIDs")
	}
	builder.WriteString("\n")

	// TODO: Include the fields if they are implemented later (uncomment and modify as needed)
	/*
		builder.WriteString(fmt.Sprintf("ShardNO: %d\n", l.ShardNO))
		builder.WriteString(fmt.Sprintf("From: %v\n", l.From))
		builder.WriteString(fmt.Sprintf("CurrentTurn: %d\n", l.CurrentTurn))
		builder.WriteString(fmt.Sprintf("Signature: %v\n", l.Signature))
	*/

	return builder.String()
}
