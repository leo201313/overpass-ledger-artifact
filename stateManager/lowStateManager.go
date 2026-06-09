package stateManager

import (
	"opl/common"
	"opl/elements"
)

// LowStateManager manage the world state and the VDT.
// Calling ProcessBlock would process the transactions and wrap them into a block
// and at the same time write on the VDT.
// StepNextEpoch would reset the current base version and nounce, also trim the VDT
type LowStateManager interface {
	StepNextEpoch(epoch elements.Epoch)
	ProcessBlock(blockID common.Hash, txs []elements.Transaction) elements.Block
}
