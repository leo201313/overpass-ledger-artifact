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

type Worker struct {
	NodeName       string   // the name of this worker
	SelfAddr       string   // the self IP address
	CdntAddr       string   // the IP address of associated coordinator
	Party          string   // the party name
	Shard          uint8    // the shard number
	ShardPeerAddrs []string // the IP addresses of other workers in the shard (cannot include self)

	HostClient *network.NodeHost // the NodeHost of this worker

	CdntPeer *network.Peer      // the Peer of associated coordinator
	ShardPG  *network.PeerGroup // the PeerGroup of the shard

	threshold0 int // used for inferring whether to step into next during the PBFT
	threshold1 int // used for inferring whether to launch another consensus round
	threshold2 int // used for inferring whether the epochs have been committed

	worldStateManager *stateManager.WorldStateManager
	lowStateManager   stateManager.LowStateManager

	stateContainer *workerStates

	txPool *workerTxPool

	Running bool
}

func NewWorker(nodeName, selfAddr, coordinatorAddr, partyName string, shardNumber uint8, shardPeerAddrs []string, wsm *stateManager.WorldStateManager, sme smartcontract.SmartContractEngine) (*Worker, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("the name of worker cannot be empty")
	}

	if len(partyName) == 0 {
		return nil, fmt.Errorf("the name of party cannot be empty")
	}

	if selfAddr == coordinatorAddr {
		return nil, fmt.Errorf("the coordinator address and self address are same, both are :%s", selfAddr)
	}

	shardPeers := make([]string, 0)
	for _, ps := range shardPeerAddrs {
		if ps == selfAddr {
			continue
		}
		shardPeers = append(shardPeers, ps)
	}

	shardAmount := len(shardPeers) + 1
	f := (shardAmount - 1) / 3
	th0 := shardAmount - f
	th1 := shardAmount
	th2 := shardAmount

	host := network.NewNodeHost(nodeName, selfAddr)
	host.Start()

	selfIndex := utils.ComputeIndexFromIPAddr(selfAddr, shardPeerAddrs)

	workerState := newWorkerStates(selfIndex, shardAmount)

	lsm := stateManager.NewSimpleLowStateManager(wsm, sme, shardNumber)

	worker := Worker{
		NodeName:          nodeName,
		SelfAddr:          selfAddr,
		CdntAddr:          coordinatorAddr,
		Party:             partyName,
		Shard:             shardNumber,
		ShardPeerAddrs:    shardPeers,
		HostClient:        host,
		CdntPeer:          nil,
		ShardPG:           network.NewPeerGroup(),
		threshold0:        th0,
		threshold1:        th1,
		threshold2:        th2,
		worldStateManager: wsm,
		lowStateManager:   lsm,
		stateContainer:    workerState,
		txPool:            newWorkerTxPool(),
	}

	return &worker, nil
}

func (w *Worker) StartDial() {
	dialTask := make([]string, 0)
	//dialTask = append(dialTask, w.CdntAddr)
	for _, p := range w.ShardPeerAddrs {
		dialTask = append(dialTask, p)
	}
	tasks := network.SimpleCreateDialTasks(w.SelfAddr, dialTask, 3, false) // for test, we use unencoded connection
	w.HostClient.Dial(tasks)
}

func (w *Worker) cdntMsgProcess(msg network.Msg) error {
	switch msg.Code {
	case SynchronizeEpochMsg:
		var uppEpochMsg UppEpochMsg
		if err := rlp.Decode(msg.Payload, &uppEpochMsg); err != nil {
			return err
		}
		err := w.stateContainer.processUppEpochMsg(uppEpochMsg)
		return err

	//case RoundTickMsg:
	//	var uppTickMsg UppTickMsg
	//	if err := rlp.Decode(msg.Payload, &uppTickMsg); err != nil {
	//		return err
	//	}
	//
	//	lmutm := LowMulticastUppTickMsg{
	//		BaseVersion: uppTickMsg.BaseVersion,
	//		NextVersion: uppTickMsg.NextVersion,
	//	}
	//
	//	err := w.ShardPG.Broadcast(RoundTickMulticastMsg, &lmutm)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	w.stateContainer.lus.Update(uppTickMsg.BaseVersion, uppTickMsg.NextVersion)

	default:
		return fmt.Errorf("unknown message with code %d is got", msg.Code)

	}
	return nil
}

func (w *Worker) shardMsgProcess(msg network.Msg) error {
	switch msg.Code {
	case TransactionMsg:
		var txMsg TxMsg
		if err := rlp.Decode(msg.Payload, &txMsg); err != nil {
			return err
		}
		w.txPool.addTx(txMsg.Content)
	case TransactionGroupMsg:
		var txsMsg TxGroupMsg
		if err := rlp.Decode(msg.Payload, &txsMsg); err != nil {
			return err
		}
		w.txPool.addTxGroup(txsMsg.Content)
	case Low_PreprepareMsg:
		var preprepareMsg LowPreprepareMsg
		if err := rlp.Decode(msg.Payload, &preprepareMsg); err != nil {
			return err
		}
		err := w.stateContainer.processLowPreprepareMsg(preprepareMsg)
		return err
	case Low_PrepareMsg:
		var prepareMsg LowPrepareMsg
		if err := rlp.Decode(msg.Payload, &prepareMsg); err != nil {
			return err
		}
		w.stateContainer.addLowPrepareMsg(prepareMsg)
	case Low_CommitMsg:
		var commitMsg LowCommitMsg
		if err := rlp.Decode(msg.Payload, &commitMsg); err != nil {
			return err
		}
		w.stateContainer.addLowCommitMsg(commitMsg)
	case Low_DoneMsg:
		var doneMsg LowDoneMsg
		if err := rlp.Decode(msg.Payload, &doneMsg); err != nil {
			return err
		}
		w.stateContainer.addLowDoneMsg(doneMsg)
	case Low_EpochStepInMsg:
		var stepInMsg LowEpochStepInMsg
		if err := rlp.Decode(msg.Payload, &stepInMsg); err != nil {
			return err
		}
		err := w.stateContainer.processEpochStepInMsg(stepInMsg)
		return err
	case Low_EpochDoneMsg:
		var epochDoneMsg LowEpochDoneMsg
		if err := rlp.Decode(msg.Payload, &epochDoneMsg); err != nil {
			return err
		}
		w.stateContainer.addLowEpochDoneMsg(epochDoneMsg)

	//case RoundTickMulticastMsg:
	//	var tickMulticastMsg LowMulticastUppTickMsg
	//	if err := rlp.Decode(msg.Payload, &tickMulticastMsg); err != nil {
	//		return err
	//	}
	//	w.stateContainer.lus.Update(tickMulticastMsg.BaseVersion, tickMulticastMsg.NextVersion)

	default:
		return fmt.Errorf("unknown message with code %d is got", msg.Code)
	}
	return nil
}

func (w *Worker) BindPeers() error {
	cdntPeer, err := w.HostClient.BindPeer(w.CdntAddr, w.cdntMsgProcess)
	if err != nil {
		return err
	}
	w.CdntPeer = cdntPeer

	for _, p := range w.ShardPeerAddrs {
		shardPeer, err := w.HostClient.BindPeer(p, w.shardMsgProcess)
		if err != nil {
			return err
		}
		w.ShardPG.AddPeer(shardPeer)
	}

	return nil
}

func (w *Worker) stateCheckLoop() {
	scanTicker := time.NewTicker(coes.WorkerStateCheckInterval)
	for {
		<-scanTicker.C
		w.Operate()
	}
}

// Run should be called after dial and bind
func (w *Worker) Run() error {
	// TODO: multiple check should be done here

	w.Running = true
	go w.stateCheckLoop()
	return nil
}

func (w *Worker) Operate() {
	w.stateContainer.stateMux.Lock()
	defer w.stateContainer.stateMux.Unlock()

	switch w.stateContainer.stage {
	case spare:
		if w.stateContainer.currentTurn == w.stateContainer.currentRank {
			// it is at the turn of leader

			if !w.stateContainer.newEpoch.isEmpty() {
				if w.stateContainer.localEpochVersion == common.INITIAL_VERSION && w.stateContainer.nextNonce != 2 {
					// this shard has not done initiation yet
					goto nonceCheck
				}

				if w.stateContainer.nextNonce == 0 {
					// this shard has not generated block yet
					goto nonceCheck
				}

				// new epoch is got and the shard should step into the next epoch
				err := w.announceEpochStep()
				if err != nil {
					panic(err)
				}
				return
			}

		nonceCheck:
			if w.stateContainer.nextNonce != 0 {
				if w.stateContainer.localEpochVersion == common.INITIAL_VERSION {
					// At the initialization,
					// the upper layer has begun a round of consensus, and the lower layer is still wait for the Epoch Commit
					// thus, the lower layer is able to launch another round of consensus
					if w.stateContainer.nextNonce == 1 {
						// only allow once!
						w.tryLaunchRound()
					}
					return
				} else {
					// the upper has not committed yet
					return
				}
			}

			w.tryLaunchRound()
			return

		} else {
			// it is at the turn of follower

			if w.stateContainer.esimHolder.occupy() {
				esim := w.stateContainer.esimHolder.retrieve()
				err := w.stateContainer.processEpochStepInMsg_SP(esim)
				if err != nil {
					panic(err)
				}
			}

			if w.stateContainer.ppmHolder.occupy() {
				ppm := w.stateContainer.ppmHolder.retrieve()
				err := w.stateContainer.processLowPreprepareMsg_SP(ppm)
				if err != nil {
					panic(err)
				}
			}
			return
		}
	case ordering_wait_transactions:
		if w.txPool.containAll(w.stateContainer.candidateTxIDs) {
			retrievedTxs := w.txPool.retrieveByIDs(w.stateContainer.candidateTxIDs)
			candidateBlock := elements.Block{
				BlockID:      common.Hash{},
				ShardNO:      w.Shard,
				Version:      w.stateContainer.localEpochVersion,
				Nonce:        w.stateContainer.nextNonce,
				Transactions: retrievedTxs,
			}
			candidateBlock.SetBlockID()

			w.stateContainer.candidateBlock = &candidateBlock

			prepareMsg := LowPrepareMsg{
				Height: w.stateContainer.nowHeight,
				Nonce:  w.stateContainer.nextNonce,
			}
			err := w.ShardPG.Broadcast(Low_PrepareMsg, &prepareMsg)
			if err != nil {
				panic(err)
			}

			w.stateContainer.stage = ordering_wait_prepare
			return
		} else {
			return
		}
	case ordering_wait_prepare:
		if w.stateContainer.IsPrepare(w.threshold0) {
			commitMsg := LowCommitMsg{
				Height: w.stateContainer.nowHeight,
				Nonce:  w.stateContainer.nextNonce,
			}
			err := w.ShardPG.Broadcast(Low_CommitMsg, &commitMsg)
			if err != nil {
				panic(err)
			}
			w.stateContainer.stage = ordering_wait_commit
		}

	case ordering_wait_prepare_leader:
		if w.stateContainer.IsPrepare_Leader(w.threshold0) {
			commitMsg := LowCommitMsg{
				Height: w.stateContainer.nowHeight,
				Nonce:  w.stateContainer.nextNonce,
			}
			err := w.ShardPG.Broadcast(Low_CommitMsg, &commitMsg)
			if err != nil {
				panic(err)
			}
			w.stateContainer.stage = ordering_wait_commit
		}

	case ordering_wait_commit:
		if w.stateContainer.IsCommit(w.threshold0) {
			// first should execute the transactions and construct a block
			tBlock := w.lowStateManager.ProcessBlock(w.stateContainer.candidateBlock.BlockID, w.stateContainer.candidateBlock.Transactions)

			// then upload the block to the upper layer
			blockMsg := LowBlockMsg{Block: tBlock}
			err := w.CdntPeer.Send(UploadBlockMsg, &blockMsg)
			if err != nil {
				panic(err)
			}

			// notify other members in the shard that this node has done
			doneMsg := LowDoneMsg{Height: w.stateContainer.nowHeight, Nonce: w.stateContainer.nextNonce}
			err = w.ShardPG.Broadcast(Low_DoneMsg, &doneMsg)
			if err != nil {
				panic(err)
			}

			w.stateContainer.stage = confirming_wait_done
		}

	case confirming_wait_done:
		if w.stateContainer.IsConfirmed(w.threshold1) {
			// empty caches
			w.stateContainer.EmptyCache(w.stateContainer.nowHeight, w.stateContainer.nextNonce)

			// step into next turn
			w.stateContainer.StepNextTurn()
			w.stateContainer.stage = spare
		}

	case committing_wait_epochs:
		if w.stateContainer.newEpoch.containEpoch(w.stateContainer.candidateEpochIDs) {
			retrievedEpochs := w.stateContainer.newEpoch.retrieveAll(w.stateContainer.candidateEpochIDs)
			prevHeight := w.stateContainer.nowHeight
			for _, epoch := range retrievedEpochs {
				if epoch.PreviousEpoch != w.stateContainer.localEpochVersion {
					panic("the epoch version is not matched")
				}
				if (w.stateContainer.nowHeight + 1) != epoch.Height {
					panic("the epoch height is not matched")
				}

				w.worldStateManager.AppendEpoch(epoch)
				w.worldStateManager.CommitStateSet(epoch.StateCommitSet)
				w.lowStateManager.StepNextEpoch(epoch)

				w.stateContainer.localEpochVersion = epoch.EpochID
				w.stateContainer.nowHeight += 1
				w.stateContainer.nextNonce = 0
			}

			ledm := LowEpochDoneMsg{
				PrevHeight: prevHeight,
				NowHeight:  w.stateContainer.nowHeight,
			}

			w.stateContainer.stage = epoch_update_wait_done
			err := w.ShardPG.Broadcast(Low_EpochDoneMsg, &ledm)
			if err != nil {
				panic(err)
			}
		}

	case epoch_update_wait_done:
		if w.stateContainer.IsEpochUpdateDone(w.threshold2) {
			w.stateContainer.EmptyEpochDoneMsg()
			w.stateContainer.stage = spare
		}

	case epoch_update_wait_done_leader:
		if w.stateContainer.IsEpochUpdateDone_Leader(w.threshold2) {
			w.stateContainer.EmptyEpochDoneMsg()
			w.stateContainer.stage = spare
		}

	default:
		err := fmt.Errorf("unknown stage %d is got", w.stateContainer.stage)
		panic(err)
	}
}

// TryLaunchRound tries to launch a new round of lower consensus if the block cache is not empty.
// If a round is launched, the stage is changed.
// NOTE: this function can only be used in Operate
func (w *Worker) tryLaunchRound() {
	if w.txPool.amount() < coes.LeastTransaction {
		// not enough transactions got, do nothing
		return
	}

	txs, txIDs := w.txPool.retrieveAll(coes.MaxTransaction)

	candidateBlock := elements.Block{
		BlockID:      common.Hash{},
		ShardNO:      w.Shard,
		Version:      w.stateContainer.localEpochVersion,
		Nonce:        w.stateContainer.nextNonce,
		Transactions: txs,
	}
	candidateBlock.SetBlockID()

	ppm := LowPreprepareMsg{
		ShardNO:    w.Shard,
		EpochID:    w.stateContainer.localEpochVersion,
		Height:     w.stateContainer.nowHeight,
		Nonce:      w.stateContainer.nextNonce,
		TxSequence: txIDs,
	}

	// change the state
	w.stateContainer.stage = ordering_wait_prepare_leader
	w.stateContainer.candidateBlock = &candidateBlock

	// broadcast the preprepare message
	err := w.ShardPG.Broadcast(Low_PreprepareMsg, &ppm)
	if err != nil {
		panic(err)
	}

}

// announceEpochStep announces that all workers in the shard
// are required to step into new epoch.
// NOTE: this function can only be used in Operate
func (w *Worker) announceEpochStep() error {
	epochs := w.stateContainer.newEpoch.selectAll()
	prevHeight := w.stateContainer.nowHeight
	epochIDs := make([]common.Hash, len(epochs))
	for i, epoch := range epochs {
		epochIDs[i] = epoch.EpochID

		if epoch.PreviousEpoch != w.stateContainer.localEpochVersion {
			return fmt.Errorf("fail in announceEpochStep, now local version is %x, the new epoch %x request previous version %x", w.stateContainer.localEpochVersion, epoch.EpochID, epoch.PreviousEpoch)
		}

		if (w.stateContainer.nowHeight + 1) != epoch.Height {
			return fmt.Errorf("fail in announceEpochStep, as expect epoch with height %d, but got height %d", w.stateContainer.nowHeight+1, epoch.Height)
		}

		w.worldStateManager.AppendEpoch(epoch)
		w.worldStateManager.CommitStateSet(epoch.StateCommitSet)
		w.lowStateManager.StepNextEpoch(epoch)

		w.stateContainer.localEpochVersion = epoch.EpochID
		w.stateContainer.nowHeight += 1
		w.stateContainer.nextNonce = 0
	}

	lesm := LowEpochStepInMsg{
		PrevHeight: prevHeight,
		NowHeight:  w.stateContainer.nowHeight,
		EpochIDs:   epochIDs,
	}

	w.stateContainer.stage = epoch_update_wait_done_leader
	// broadcast the epoch step in message
	err := w.ShardPG.Broadcast(Low_EpochStepInMsg, &lesm)
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) GetTransaction(tx elements.Transaction) error {
	w.txPool.addTx(tx)
	txMsg := TxMsg{Content: tx}
	err := w.ShardPG.Broadcast(TransactionMsg, &txMsg)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) GetTransactionGroup(txs []elements.Transaction) error {
	w.txPool.addTxGroup(txs)
	txsMsg := TxGroupMsg{Content: txs}
	err := w.ShardPG.Broadcast(TransactionGroupMsg, &txsMsg)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) BackCurrentState() (stageCode uint64, nowVersion common.Hash, nowHeight uint64) {
	w.stateContainer.stateMux.Lock()
	defer w.stateContainer.stateMux.Unlock()
	return uint64(w.stateContainer.stage), w.stateContainer.localEpochVersion, w.stateContainer.nowHeight
}

func (w *Worker) GetTxIDsInEpochByHeight(height int) (txIDs []common.Hash, have bool) {
	txIDs, err := w.worldStateManager.GetTxIDsInEpochByHeight(uint64(height))
	if err != nil {
		return nil, false
	}
	return txIDs, true
}

func (w *Worker) GetTxIDsAndTypesInEpochByHeight(height int) (txIDs []common.Hash, txTypes []uint8, have bool) {
	txIDs, txTypes, err := w.worldStateManager.GetTxIDsAndTypesInEpochByHeight(uint64(height))
	if err != nil {
		return nil, nil, false
	}
	return txIDs, txTypes, true
}

func (w *Worker) BackAllStateInfo() string {
	w.stateContainer.stateMux.Lock()
	defer w.stateContainer.stateMux.Unlock()

	res := w.stateContainer.String()
	res += "\n"
	res += w.txPool.String()
	return res
}

// CreatAccounts is only used for test.
// Creat accountNum accounts in the database
func (w *Worker) CreatAccounts(accountNum int) {
	for i := 0; i < accountNum; i++ {
		accountName := smartcontract.BackAccountNameByIndex(i)
		accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountName)))
		w.worldStateManager.WriteState(accountAddr, smartcontract.InitValueBytes)
	}
}
