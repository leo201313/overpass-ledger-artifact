package entity

import (
	"crypto/sha256"
	"fmt"
	"opl/coes"
	"opl/common"
	"opl/elements"
	"opl/network"
	"opl/rlp"
	"opl/smartcontract"
	"opl/stateManager"
	"opl/utils"
	"time"
)

type Coordinator struct {
	NodeName       string   // the name of this coordinator
	SelfAddr       string   // the self IP address
	Party          string   // the party name
	CdntAddrs      []string // the IP addresses of other coordinators (cannot include self)
	WorkerAddrs    []string // the IP addresses of connected workers
	WorkerShardNOs []uint8  // should be with the same sequence of worker addr

	HostClient *network.NodeHost // the NodeHost of this coordinator

	CdntPG   *network.PeerGroup // the PeerGroup of the coordination communication group
	WorkerPG *network.PeerGroup // the PeerGroup of the workers associated with this coordinator of the same party

	threshold0 int // used for PBFT
	threshold1 int // used for done synchronization

	blockManager *blockCache // manage all blocks uploaded from the lower layer

	worldStateManager *stateManager.WorldStateManager
	uppStateManager   stateManager.UppStateManager
	stateContainer    *cdntState // containing all states about consensus

	//testMode      bool
	//testContainer *testStateContainer
	Running bool
}

//type testStateContainer struct {
//	loopTrigger  bool
//	triggerClose chan struct{}
//
//	managerAddr       string
//	managerPeer       *network.Peer
//	nowTrigger        bool
//	triggerStartTime  time.Time
//	finishTime        time.Time
//	lastEpochTxNum    int
//	lastEpochBlockNum int
//}
//
//func newTestStateContainer(managerAddr string) *testStateContainer {
//	return &testStateContainer{
//		managerAddr:       managerAddr,
//		managerPeer:       nil,
//		nowTrigger:        false,
//		triggerStartTime:  time.Time{},
//		finishTime:        time.Time{},
//		lastEpochTxNum:    0,
//		lastEpochBlockNum: 0,
//	}
//}

func NewCoordinator(nodeName, selfAddr, partyName string, cdntAddrs, workerAddrs []string, shardNOs []uint8, wsm *stateManager.WorldStateManager, sme smartcontract.SmartContractEngine) (*Coordinator, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("the name of coordinator cannot be empty")
	}

	if len(partyName) == 0 {
		return nil, fmt.Errorf("the name of party cannot be empty")
	}

	cdntPeers := make([]string, 0)
	for _, cp := range cdntAddrs {
		if cp == selfAddr {
			continue
		}
		cdntPeers = append(cdntPeers, cp)
	}

	if len(workerAddrs) != len(shardNOs) {
		return nil, fmt.Errorf("the number of workers' addresses and shard NOs are not matched")
	}

	workerPeers := make([]string, 0)
	workerShardNOs := make([]uint8, 0)

	for i := 0; i < len(workerAddrs); i++ {
		if workerAddrs[i] == selfAddr {
			continue
		}
		workerPeers = append(workerPeers, workerAddrs[i])
		workerShardNOs = append(workerShardNOs, shardNOs[i])
	}

	cdntAmount := len(cdntPeers) + 1
	f := (cdntAmount - 1) / 3
	th0 := cdntAmount - f
	th1 := cdntAmount

	selfRank := utils.ComputeIndexFromIPAddr(selfAddr, cdntPeers)

	host := network.NewNodeHost(nodeName, selfAddr)
	host.Start()

	uppStateManager := stateManager.NewSimpleUppStateManager(sme, wsm)

	cdnt := Coordinator{
		NodeName:          nodeName,
		SelfAddr:          selfAddr,
		Party:             partyName,
		CdntAddrs:         cdntPeers,
		WorkerAddrs:       workerPeers,
		WorkerShardNOs:    workerShardNOs,
		HostClient:        host,
		CdntPG:            network.NewPeerGroup(),
		WorkerPG:          network.NewPeerGroup(),
		threshold0:        th0,
		threshold1:        th1,
		stateContainer:    newCdntState(selfRank, cdntAmount),
		blockManager:      nil,
		worldStateManager: wsm,
		uppStateManager:   uppStateManager,
	}

	return &cdnt, nil
}

func (c *Coordinator) StartDial() {
	tasks1 := network.SimpleCreateDialTasks(c.SelfAddr, c.CdntAddrs, 3, false) // for test, we use unencoded connection
	tasks2 := network.SimpleCreateDialTasks_All(c.SelfAddr, c.WorkerAddrs, 3, false)
	tasks := append(tasks1, tasks2...)
	c.HostClient.Dial(tasks)
}

func (c *Coordinator) BindPeers() error {
	for _, p := range c.CdntAddrs {
		cdntPeer, err := c.HostClient.BindPeer(p, c.handleCdntMsg)
		if err != nil {
			return err
		}
		c.CdntPG.AddPeer(cdntPeer)
	}

	for _, p := range c.WorkerAddrs {
		workerPeer, err := c.HostClient.BindPeer(p, c.handleWorkerMsg)
		if err != nil {
			return err
		}
		c.WorkerPG.AddPeer(workerPeer)
	}

	//if c.testMode {
	//	managerPeer, err := c.HostClient.BindPeer(c.testContainer.managerAddr, c.handleManagerMsg)
	//	if err != nil {
	//		return err
	//	}
	//	c.testContainer.managerPeer = managerPeer
	//}

	return nil
}

// Run should be called after dial and bind
func (c *Coordinator) Run() error {
	// TODO: multiple check should be done here

	c.blockManager = newBlockCache(c.WorkerShardNOs)
	c.Running = true
	go c.stateCheckLoop()

	return nil

}

//
//func (c *Coordinator) handleManagerMsg(msg network.Msg) error {
//	switch msg.Code {
//	case Test_BlocksMsg:
//		var testBlocksMsg TestBlocksUploadMsg
//		if err := rlp.Decode(msg.Payload, &testBlocksMsg); err != nil {
//			return err
//		}
//		for i := 0; i < len(testBlocksMsg.Blocks); i++ {
//			c.blockManager.addBlock(testBlocksMsg.Blocks[i])
//		}
//
//	case Test_StateQueryMsg:
//		nowState := uint8(c.stateContainer.stage) // here no need to lock the stateContainer
//		deltaTime := c.testContainer.finishTime.Sub(c.testContainer.triggerStartTime)
//		timeUsed := uint64(deltaTime.Microseconds())
//		txNum := uint64(c.testContainer.lastEpochTxNum)
//		blockNum := uint64(c.testContainer.lastEpochBlockNum)
//
//		testStateMsg := TestStateMsg{
//			State:    nowState,
//			TimeUsed: timeUsed,
//			TxNum:    txNum,
//			BlockNum: blockNum,
//		}
//
//		err := c.testContainer.managerPeer.Send(Test_StateBackMsg, &testStateMsg)
//		if err != nil {
//			return err
//		}
//
//	case Test_TriggerMsg:
//		c.TryLaunchEpoch() // has no need to handle locks
//	case Test_LoopTiggerMsg:
//		go c.loopTrigger()
//	case Test_StopLoopTriggerMsg:
//		c.closeLoopTrigger()
//	}
//
//	return nil
//}

func (c *Coordinator) handleCdntMsg(msg network.Msg) error {
	switch msg.Code {
	case Upp_PreprepareMsg:
		var preprepareMsg UppPreprepareMsg
		if err := rlp.Decode(msg.Payload, &preprepareMsg); err != nil {
			return err
		}

		err := c.stateContainer.processUppPreprepareMsg(preprepareMsg)
		return err

	case Upp_PrepareMsg:
		var prepareMsg UppPrepareMsg
		if err := rlp.Decode(msg.Payload, &prepareMsg); err != nil {
			return err
		}
		c.stateContainer.addUppPrepareMsg(prepareMsg)

	case Upp_CommitMsg:
		var commitMsg UppCommitMsg
		if err := rlp.Decode(msg.Payload, &commitMsg); err != nil {
			return err
		}
		c.stateContainer.addUppCommitMsg(commitMsg)

	case Upp_DoneMsg:
		var doneMsg UppDoneMsg
		if err := rlp.Decode(msg.Payload, &doneMsg); err != nil {
			return err
		}
		c.stateContainer.addUppDoneMsg(doneMsg)

	default:
		return fmt.Errorf("unknown message with code %d is got", msg.Code)
	}
	return nil
}

func (c *Coordinator) handleWorkerMsg(msg network.Msg) error {
	switch msg.Code {
	case UploadBlockMsg:
		var lowBlockMsg LowBlockMsg
		if err := rlp.Decode(msg.Payload, &lowBlockMsg); err != nil {
			return err
		}
		if err := c.blockManager.addBlock(lowBlockMsg.Block); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown message with code %d is got", msg.Code)
	}
	return nil
}

func (c *Coordinator) stateCheckLoop() {
	scanTicker := time.NewTicker(coes.StateCheckInterval)
	for {
		<-scanTicker.C
		c.Operate()
	}
}

func (c *Coordinator) Operate() {
	c.stateContainer.stageMux.Lock()
	defer c.stateContainer.stageMux.Unlock()

	switch c.stateContainer.stage {
	case stage_spare:
		if c.stateContainer.currentTurn == c.stateContainer.selfRank {
			// be the leader
			c.TryLaunchEpoch()
			return
		} else {
			// be the follower, do nothing
			return
		}
		return
	case stage_ordering_wait_blocks:
		if c.blockManager.findAll(c.stateContainer.candidateMap) {
			groupedBlocks := c.blockManager.retrieveAll(c.stateContainer.candidateMap)
			blocks, err := AggregateSequence(groupedBlocks, c.stateContainer.candidateBlockSequence, c.stateContainer.candidateShardAssociated)
			if err != nil {
				panic(err)
			}
			c.stateContainer.candidateBlocks = blocks

			prepareMsg := UppPrepareMsg{EpochID: c.stateContainer.nextVersion}
			err = c.CdntPG.Broadcast(Upp_PrepareMsg, &prepareMsg)
			if err != nil {
				panic(err)
			}

			c.stateContainer.stage = stage_ordering_wait_prepare

			return
		} else {
			return
		}

	case stage_ordering_wait_prepare:
		if c.stateContainer.IsPrepare(c.threshold0) {

			commitMsg := UppCommitMsg{EpochID: c.stateContainer.nextVersion}
			err := c.CdntPG.Broadcast(Upp_CommitMsg, &commitMsg)
			if err != nil {
				panic(err)
			}
			c.stateContainer.stage = stage_ordering_wait_commit
		}

	case stage_ordering_wait_prepare_leader:
		if c.stateContainer.IsPrepare_Leader(c.threshold0) {

			commitMsg := UppCommitMsg{EpochID: c.stateContainer.nextVersion}
			err := c.CdntPG.Broadcast(Upp_CommitMsg, &commitMsg)
			if err != nil {
				panic(err)
			}
			c.stateContainer.stage = stage_ordering_wait_commit
		}

	case stage_ordering_wait_commit:
		if c.stateContainer.IsCommit(c.threshold0) {
			c.ProcessEpoch()

			doneMsg := UppDoneMsg{EpochID: c.stateContainer.nextVersion}
			err := c.CdntPG.Broadcast(Upp_DoneMsg, &doneMsg)
			if err != nil {
				panic(err)
			}
			c.stateContainer.stage = stage_confirming_wait
		}

	case stage_confirming_wait:
		if c.stateContainer.IsConfirmed(c.threshold1) {
			err := c.CommitEpoch()
			if err != nil {
				panic(err)
			}

			c.stateContainer.Complete()

			c.stateContainer.stage = stage_spare
		}

	case stage_confiming_wait_but_new_round_begin:
		if c.stateContainer.IsConfirmed_Variant(c.threshold1) {
			err := c.CommitEpoch()
			if err != nil {
				panic(err)
			}

			c.stateContainer.Complete_Variant()

			c.stateContainer.stage = stage_ordering_wait_blocks
		}
	default:
		err := fmt.Errorf("unknown stage %d is got", c.stateContainer.stage)
		panic(err)
	}

}

func (c *Coordinator) ProcessEpoch() {
	epoch := elements.Epoch{
		EpochID:                 c.stateContainer.nextVersion,
		PreviousEpoch:           c.stateContainer.nowVersion,
		Height:                  c.stateContainer.nowHeight + 1,
		BlockSequence:           c.stateContainer.candidateBlockSequence,
		BlockAssociatedShardNOs: c.stateContainer.candidateShardAssociated,
		Receipts:                nil,
		StateCommitSet:          nil,
	}

	c.uppStateManager.CommitNowVCT()
	c.uppStateManager.StepNextEpoch(epoch.EpochID)
	receipts := c.uppStateManager.ProcessEpoch(c.stateContainer.candidateBlocks)
	stateCommitSet := c.uppStateManager.ExportStateCommitSet()

	epoch.Receipts = receipts
	epoch.StateCommitSet = stateCommitSet

	c.stateContainer.nowEpoch = &epoch
}

func (c *Coordinator) CommitEpoch() error {
	// First announce the epoch update message
	uppEpochMsg := UppEpochMsg{Epoch: *c.stateContainer.nowEpoch}
	err := c.WorkerPG.Broadcast(SynchronizeEpochMsg, &uppEpochMsg)
	if err != nil {
		return err
	}

	c.worldStateManager.AppendEpoch(*c.stateContainer.nowEpoch)
	c.worldStateManager.CommitStateSet(c.stateContainer.nowEpoch.StateCommitSet)

	//if c.testMode {
	//	testEpochCommitMsg := TestEpochCommitMsg{Epoch: *c.stateContainer.nowEpoch}
	//	if c.testContainer.nowTrigger {
	//		c.testContainer.nowTrigger = false
	//		c.testContainer.finishTime = time.Now()
	//		err := c.testContainer.managerPeer.Send(Test_EpochMsg, &testEpochCommitMsg)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}

	return nil
}

// Check whether the state stage is at state_spare. If it is, then
// launch a new epoch and change the stage into stage_ordering_wait_prepare_leader
// NOTE: TryLaunchEpoch now is only used in Coordinator.Operate, and has no need to
// check the stage
func (c *Coordinator) TryLaunchEpoch() {
	//c.stateContainer.stageMux.Lock()
	//defer c.stateContainer.stageMux.Unlock()
	//if c.stateContainer.stage != stage_spare {
	//	return
	//}

	blockMap, ready := c.blockManager.selectAll_SP()
	if !ready {
		return
	}

	orderedBlocks, sharNOs := EqualOrderBlocks(blockMap)

	previousVersion := c.stateContainer.nowVersion
	nextHeight := c.stateContainer.nowHeight + 1

	nextEpoch := elements.Epoch{
		EpochID:                 common.Hash{},
		PreviousEpoch:           previousVersion,
		Height:                  nextHeight,
		BlockSequence:           orderedBlocks,
		BlockAssociatedShardNOs: sharNOs,
		Receipts:                nil,
		StateCommitSet:          nil,
	}

	nextVersion := nextEpoch.Hash()

	ppm := UppPreprepareMsg{
		EpochID:         nextVersion,
		PreviousVersion: previousVersion,
		Height:          nextHeight,
		BlockSequence:   orderedBlocks,
		BlockAssociated: sharNOs,
	}

	// update self
	c.stateContainer.nextVersion = nextVersion
	c.stateContainer.candidateBlockSequence = orderedBlocks
	c.stateContainer.candidateShardAssociated = sharNOs
	c.stateContainer.candidateMap = blockMap

	groupedBlocks := c.blockManager.retrieveAll(blockMap)
	blocks, err := AggregateSequence(groupedBlocks, orderedBlocks, sharNOs)
	if err != nil {
		panic(err)
	}
	c.stateContainer.candidateBlocks = blocks

	c.stateContainer.stage = stage_ordering_wait_prepare_leader

	//if c.testMode {
	//	c.testContainer.nowTrigger = true
	//	c.testContainer.triggerStartTime = time.Now()
	//	txAmount := 0
	//	for _, block := range blocks {
	//		txAmount += len(block.Transactions)
	//	}
	//	c.testContainer.lastEpochTxNum = txAmount
	//	c.testContainer.lastEpochBlockNum = len(blocks)
	//}

	// broadcast preprepare message
	err = c.CdntPG.Broadcast(Upp_PreprepareMsg, &ppm)
	if err != nil {
		panic(err)
	}

	//if previousVersion == common.INITIAL_VERSION {
	//	upm := UppTickMsg{
	//		BaseVersion: previousVersion,
	//		NextVersion: nextVersion,
	//	}
	//
	//	err = c.WorkerPG.Broadcast(RoundTickMsg, &upm)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
}

// IsSpare is used for test
func (c *Coordinator) IsSpare() bool {
	c.stateContainer.stageMux.Lock()
	defer c.stateContainer.stageMux.Unlock()
	if c.stateContainer.stage == stage_spare {
		return true
	}
	return false
}

// EqualOrderBlocks orders the blocks base on the shard numbers equally.
// maybe more efficient order method can be designed, but we use this equal method at this version
func EqualOrderBlocks(blockMap map[uint8][]common.Hash) ([]common.Hash, []uint8) {
	index := 0

	orderedBlocks := []common.Hash{}
	shardNOs := []uint8{}

	for {
		flag := false
		for shardNO, blockGroup := range blockMap {
			if len(blockGroup) > index {
				orderedBlocks = append(orderedBlocks, blockGroup[index])
				shardNOs = append(shardNOs, shardNO)
				flag = true
			}
		}
		if !flag {
			break
		} else {
			index += 1
		}
	}

	return orderedBlocks, shardNOs
}

//// loopTrigger is only used in test mode. It will try to launch the new consensus round
//// every certain interval.
//func (c *Coordinator) loopTrigger() {
//	if !c.testMode {
//		return
//	}
//
//	if c.testContainer.loopTrigger { // a loop trigger has been launched
//		return
//	}
//
//	c.testContainer.triggerClose = make(chan struct{})
//	triggerTicker := time.NewTicker(coes.TriggerScanInterval)
//
//	for {
//		select {
//		case <-triggerTicker.C:
//			if c.stateContainer.stage == stage_spare { // NOTE: here is no need to lock the state
//				c.TryLaunchEpoch()
//			}
//
//		case <-c.testContainer.triggerClose:
//			return
//		}
//	}
//}

//
//// closeLoopTrigger is used to close the loop trigger
//func (c *Coordinator) closeLoopTrigger() {
//	if !c.testMode {
//		return
//	}
//	if !c.testContainer.loopTrigger {
//		return
//	}
//
//	c.testContainer.triggerClose <- struct{}{}
//	c.testContainer.loopTrigger = false
//	return
//}

//// NewTestCoordinator is only used for test mode, it only connects with coordinators and test manager.
//func NewTestCoordinator(nodeName, selfAddr, partyName string, cdntAddrs []string, shardNOs []uint8, db *database.SimpleLDB, managerAddr string) *Coordinator {
//	cdntPeers := make([]string, 0)
//	for _, cp := range cdntAddrs {
//		if cp == selfAddr {
//			continue
//		}
//		cdntPeers = append(cdntPeers, cp)
//	}
//
//	cdntAmount := len(cdntPeers) + 1
//	f := (cdntAmount - 1) / 3
//	th0 := cdntAmount - f
//	th1 := cdntAmount
//
//	host := network.NewNodeHost(nodeName, selfAddr)
//	host.Start()
//
//	wsm := stateManager.NewWorldStateManager(db)
//	sme := smartcontract.NewDemoSME()
//	usm := stateManager.NewSimpleUppStateManager(sme, wsm)
//
//	cdnt := Coordinator{
//		NodeName:          nodeName,
//		SelfAddr:          selfAddr,
//		Party:             partyName,
//		CdntAddrs:         cdntPeers,
//		WorkerAddrs:       nil,
//		WorkerShardNOs:    shardNOs,
//		HostClient:        host,
//		CdntPG:            network.NewPeerGroup(),
//		WorkerPG:          network.NewPeerGroup(),
//		threshold0:        th0,
//		threshold1:        th1,
//		stateContainer:    newCdntState(),
//		blockManager:      nil,
//		uppStateManager:   usm,
//		worldStateManager: wsm,
//		testMode:          true,
//		testContainer:     newTestStateContainer(managerAddr),
//	}
//	return &cdnt
//}

func (cdnt *Coordinator) BackCurrentState() (stageCode uint64, nowVersion common.Hash, nowHeight uint64) {
	cdnt.stateContainer.stageMux.Lock()
	defer cdnt.stateContainer.stageMux.Unlock()
	return uint64(cdnt.stateContainer.stage), cdnt.stateContainer.nowVersion, cdnt.stateContainer.nowHeight
}

// BackAllStateInfo for debug use
func (cdnt *Coordinator) BackAllStateInfo() string {
	cdnt.stateContainer.stageMux.Lock()
	defer cdnt.stateContainer.stageMux.Unlock()
	res := cdnt.stateContainer.String()
	res += "\n"
	res += cdnt.blockManager.String()

	return res
}

// CreatAccounts is only used for test.
// Creat accountNum accounts in the database
func (cdnt *Coordinator) CreatAccounts(accountNum int) {
	for i := 0; i < accountNum; i++ {
		accountName := smartcontract.BackAccountNameByIndex(i)
		accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountName)))
		cdnt.worldStateManager.WriteState(accountAddr, smartcontract.InitValueBytes)
	}
}
