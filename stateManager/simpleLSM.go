package stateManager

import (
	"opl/common"
	"opl/elements"
	"opl/smartcontract"
)

type SimpleLowStateManager struct {
	baseVersion common.Hash
	nonce       uint64 // here we make all workers maintain the nonces themselves

	VDT     *VersionDependencyTable
	ShardNo uint8

	wsm *WorldStateManager
	sme smartcontract.SmartContractEngine
}

func NewSimpleLowStateManager(wsm *WorldStateManager, sme smartcontract.SmartContractEngine, shardNo uint8) *SimpleLowStateManager {
	lsm := &SimpleLowStateManager{
		baseVersion: common.INITIAL_VERSION,
		nonce:       0,
		VDT:         NewVDT(),
		ShardNo:     shardNo,
		wsm:         wsm,
		sme:         sme,
	}

	sme.LowInstallReadWriteFunc(lsm.ReadState, lsm.WriteState)
	return lsm
}

func (lsm *SimpleLowStateManager) StepNextEpoch(epoch elements.Epoch) {
	lsm.baseVersion = epoch.EpochID
	lsm.nonce = 0

	//// As now the contract only contain k-v store and smallbank, this trim strategy
	//// is efficient enough. More comprehensive trim strategy maybe implemented in the future.
	//for i := 0; i < len(epoch.StateCommitSet); i++ {
	//	if epoch.StateCommitSet[i].Version == epoch.EpochID {
	//		lsm.VDT.TrimState(true, epoch.StateCommitSet[i].Address, epoch.StateCommitSet[i].Version)
	//	} else {
	//		lsm.VDT.TrimState(false, epoch.StateCommitSet[i].Address, epoch.StateCommitSet[i].Version)
	//	}
	//}

	lsm.VDT.TrimState(epoch.StateCommitSet, epoch.EpochID)

}

func (lsm *SimpleLowStateManager) ReadState(addr common.Address) (have bool, value []byte, version common.Hash, versionType uint8) {
	have, value, version = lsm.VDT.ReadState(addr)
	if have {
		return have, value, version, elements.BlockVersion
	}

	have, value = lsm.wsm.ReadState(addr)

	return have, value, lsm.baseVersion, elements.EpochVersion
}

func (lsm *SimpleLowStateManager) WriteState(addr common.Address, value []byte) {
	lsm.VDT.WriteState(addr, value, lsm.VDT.NowBlockID)
}

func (lsm *SimpleLowStateManager) ProcessBlock(blockID common.Hash, txs []elements.Transaction) elements.Block {
	lsm.VDT.UpdateCurrentBlockID(blockID)
	for i := 0; i < len(txs); i++ {
		rset, wset, res := lsm.sme.LowExecuteTransaction(txs[i])
		txs[i].StateReadSet = rset
		txs[i].StateWriteSet = wset
		txs[i].Results = res
	}

	backBlock := elements.Block{
		BlockID:      blockID,
		ShardNO:      lsm.ShardNo,
		Version:      lsm.baseVersion,
		Nonce:        lsm.nonce,
		Transactions: txs,
	}

	lsm.nonce += 1

	return backBlock
}
