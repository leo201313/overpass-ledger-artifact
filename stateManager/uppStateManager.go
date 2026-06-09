package stateManager

import (
	"opl/common"
	"opl/elements"
)

// UppStateManager should first process the blocks contrained in an Epoch and results in
// a block receipt list for confirming with each other.
// After confirming, the epoch may be committed. Thus, the state commit set should be exported
// by ExportStateCommitSet to construct commit message for lower layer workers.
// For coordinator, it should commit the current VCT by using CommitNowVCT and ready for
// processing next epoch.
type UppStateManager interface {
	StepNextEpoch(epoch common.Hash)
	ProcessEpoch(blocks []elements.Block) []elements.BlockReceipt
	ExportStateCommitSet() []elements.StateCommit
	CommitNowVCT()
}
