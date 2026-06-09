package entity

// message codes for implementing the protocol of each shard
const (
	Offset_0 = iota + 0x10
	TransactionMsg
	TransactionGroupMsg
	Low_EpochStepInMsg
	Low_EpochDoneMsg

	Low_PreprepareMsg
	Low_PrepareMsg
	Low_CommitMsg
	Low_DoneMsg
)

// message codes for implementing the protocol between two layers
const (
	Offset_1            = iota + 0x30
	UploadBlockMsg      // upload block from worker to coordinator
	SynchronizeEpochMsg // commit epoch

	RoundTickMsg
	RoundTickMulticastMsg
)

// message codes for implementing the protocol among the coordinators
const (
	Offset_2 = iota + 0x50
	Upp_PreprepareMsg
	Upp_PrepareMsg
	Upp_CommitMsg
	Upp_DoneMsg
)

// message codes used between coordinators and test TestManagers
const (
	Offset_3 = iota + 0x70
	Test_BlocksMsg
	Test_EpochMsg
	Test_StateQueryMsg
	Test_StateBackMsg
	Test_TriggerMsg

	Test_LoopTiggerMsg
	Test_StopLoopTriggerMsg
)
