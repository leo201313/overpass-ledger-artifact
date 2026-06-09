package stateManager

import (
	"opl/common"
	"opl/elements"
	"opl/smartcontract"
)

const maxVCTs = 2

type SimpleUppStateManager struct {
	VCTs    [maxVCTs]*VersionCommitmentTable
	NowRank int
	wsm     *WorldStateManager
	sme     smartcontract.SmartContractEngine
}

func NewSimpleUppStateManager(sme smartcontract.SmartContractEngine, wsm *WorldStateManager) *SimpleUppStateManager {
	VCTs := [maxVCTs]*VersionCommitmentTable{}
	for i := 0; i < maxVCTs; i++ {
		VCTs[i] = NewVCT(common.INITIAL_VERSION)
	}

	usm := &SimpleUppStateManager{
		VCTs:    VCTs,
		NowRank: 0,
		wsm:     wsm,
		sme:     sme,
	}
	sme.UppInstallReadWriteFunc(usm.ReadState, usm.WriteState)

	return usm
}

func (usm *SimpleUppStateManager) StepNextEpoch(epoch common.Hash) {
	usm.VCTs[usm.NowRank] = &VersionCommitmentTable{
		BaseVersion: epoch,
		Content:     make(map[common.Address]VCTrow),
	}
}

func (usm *SimpleUppStateManager) ProcessEpoch(blocks []elements.Block) []elements.BlockReceipt {
	receipts := make([]elements.BlockReceipt, len(blocks))
	for i, block := range blocks {
		blockReceipt := elements.BlockReceipt{
			BlockID:   block.BlockID,
			TXIDs:     make([]common.Hash, len(block.Transactions)),
			TXProcess: make([]uint8, len(block.Transactions)),
			TXResults: make([][]byte, len(block.Transactions)),
		}
		for j, tx := range block.Transactions {
			processType, result := usm.processTransaction(tx, block.BlockID)
			blockReceipt.TXIDs[j] = tx.TxID
			blockReceipt.TXProcess[j] = processType
			blockReceipt.TXResults[j] = result
		}
		receipts[i] = blockReceipt
	}
	return receipts
}

func (usm *SimpleUppStateManager) ExportStateCommitSet() []elements.StateCommit {
	return usm.VCTs[usm.NowRank].ExportStates()
}

func (usm *SimpleUppStateManager) CommitNowVCT() {
	usm.NowRank = (usm.NowRank + 1) % maxVCTs
	usm.VCTs[usm.NowRank] = nil // free the VCT with old version
}

func (usm *SimpleUppStateManager) ReadState(addr common.Address) (bool, []byte) {
	have, value, _, _ := usm.VCTs[usm.NowRank].ReadState(addr)
	if have {
		return true, value
	}
	//time1 := time.Now()
	have, value = usm.wsm.ReadState(addr)
	//time.Sleep(30 * time.Microsecond) // mock as read from a huge database

	//time2 := time.Now()
	//deltatime := time2.Sub(time1).Microseconds()
	//fmt.Print(deltatime)
	return have, value
}

func (usm *SimpleUppStateManager) WriteState(addr common.Address, value []byte) {
	version := usm.VCTs[usm.NowRank].BaseVersion
	usm.VCTs[usm.NowRank].WriteState(addr, value, version, elements.EpochVersion)
}

func (usm *SimpleUppStateManager) processTransaction(Tx elements.Transaction, nowVersion common.Hash) (processType uint8, result []byte) {
	inheritAval, preloaded := usm.ValidateTxStateReadSet(Tx.StateReadSet)
	if inheritAval {
		usm.inheritTxStateWriteSet(Tx.StateWriteSet, nowVersion)
		return 0, Tx.Results
	} else {
		result = usm.sme.UppExecuteTransaction(Tx, preloaded)
		return 1, result
	}
}

func (usm *SimpleUppStateManager) ValidateTxStateReadSet(readSet []elements.StateRead) (bool, []elements.StateRead) {
	preLoaded := make([]elements.StateRead, len(readSet))
	allRight := true

outer:
	for i, state := range readSet {
		//if len(state.Value) == 0 {
		//	panic("here it is!")
		//}
		for j := 0; j < maxVCTs; j++ {
			index := (usm.NowRank - j + maxVCTs) % maxVCTs
			have, val, nowVersion, versionType := usm.VCTs[index].ReadState(state.Address)
			if !have {
				// this VCT not have
				continue
			} else {
				// this VCT got
				if nowVersion == state.Version {
					// Version match, then check next state
					preLoaded[i] = state
					continue outer
				} else {
					// Version not match, construct the latest readSet, and go outer
					allRight = false
					revisedRS := elements.StateRead{
						Address:     state.Address,
						Version:     nowVersion,
						VersionType: versionType,
						Value:       val,
					}
					preLoaded[i] = revisedRS
					continue outer
				}
			}
		}
		// All VCT not got,
		if state.VersionType == elements.EpochVersion {
			// the state is read from world states
			preLoaded[i] = state
		} else {
			// the state is read from unstable state, and the state
			// is even not in VCTs, should read from the world state
			_, val := usm.ReadState(state.Address)

			tempRS := elements.StateRead{
				Address:     state.Address,
				Version:     usm.VCTs[usm.NowRank].BaseVersion,
				VersionType: elements.EpochVersion,
				Value:       val,
			}

			preLoaded[i] = tempRS
		}
	}
	return allRight, preLoaded
}

func (usm *SimpleUppStateManager) inheritTxStateWriteSet(writeSet []elements.StateWrite, nowVersion common.Hash) {
	for _, state := range writeSet {
		usm.VCTs[usm.NowRank].WriteState(state.Address, state.Value, nowVersion, elements.BlockVersion)
	}
}
