package entity

import (
	"fmt"
	"opl/coes"
	"opl/common"
	"opl/database"
	"opl/elements"
	"opl/network"
	"opl/rlp"
	"opl/smartcontract"
	"opl/stateManager"
	"sync"
	"time"
)

// TestManager plays the role as mock worker to process lower layer blocks
// and can be only used in test, especially stressing test.
type TestManager struct {
	SelfAddr  string // the self IP address
	CdntAddrs []string
	CdntPG    *network.PeerGroup

	ShardNOs []uint8
	pools    map[uint8]*transactionPool

	HostClient *network.NodeHost

	blockManager      *blockPost
	worldStateManager *stateManager.WorldStateManager
	wsmMux            sync.Mutex // used when commit epoch and generate blocks
	agents            map[uint8]*shardAgent

	localUsedSME smartcontract.SmartContractEngine

	savedStateMessage []TestStateMsg
}

type blockPost struct {
	mux    sync.Mutex
	blocks []elements.Block
}

func newBlockPost() *blockPost {
	return &blockPost{blocks: make([]elements.Block, 0)}
}

func (bp *blockPost) add(block elements.Block) {
	bp.mux.Lock()
	defer bp.mux.Unlock()
	bp.blocks = append(bp.blocks, block)
}

func (bp *blockPost) retrieve() []elements.Block {
	bp.mux.Lock()
	defer bp.mux.Unlock()
	retrieved := bp.blocks
	bp.blocks = make([]elements.Block, 0)
	return retrieved
}

type shardAgent struct {
	shardNo uint8

	worldStateManager *stateManager.WorldStateManager
	lowStateManager   stateManager.LowStateManager

	baseVersion common.Hash
	height      uint64
	nonce       uint64
}

type transactionPool struct {
	mux     sync.Mutex
	shardNO uint8
	txs     []elements.Transaction
}

func newTransactionPool(shardNO uint8) *transactionPool {
	return &transactionPool{
		mux:     sync.Mutex{},
		shardNO: shardNO,
		txs:     make([]elements.Transaction, 0),
	}
}

func (tp *transactionPool) addTx(tx elements.Transaction) {
	tp.mux.Lock()
	defer tp.mux.Unlock()
	tp.txs = append(tp.txs, tx)
}

func (tp *transactionPool) retrieveAll() []elements.Transaction {
	tp.mux.Lock()
	defer tp.mux.Unlock()
	back := tp.txs
	tp.txs = make([]elements.Transaction, 0)
	return back
}

func newShardAgent(shardNO uint8, wsm *stateManager.WorldStateManager) *shardAgent {
	sme := smartcontract.NewDemoSME()
	lsm := stateManager.NewSimpleLowStateManager(wsm, sme, shardNO)
	return &shardAgent{
		shardNo:           shardNO,
		worldStateManager: wsm,
		lowStateManager:   lsm,
		baseVersion:       common.INITIAL_VERSION,
		height:            0,
		nonce:             0,
	}
}

func (sa *shardAgent) CommitEpoch(epoch elements.Epoch) {
	sa.baseVersion = epoch.EpochID
	sa.height = epoch.Height
	sa.nonce = 0

	//sa.worldStateManager.CommitStateSet(epoch.StateCommitSet)
	sa.lowStateManager.StepNextEpoch(epoch)
}

func (sa *shardAgent) generateBlock(txs []elements.Transaction) elements.Block {
	tempBlock := elements.Block{
		BlockID:      common.Hash{},
		ShardNO:      sa.shardNo,
		Version:      sa.baseVersion,
		Nonce:        sa.nonce,
		Transactions: nil,
	}

	block := sa.lowStateManager.ProcessBlock(tempBlock.Hash(), txs)
	sa.nonce += 1
	return block
}

func NewTestManager(selfAddr string, cdntAddrs []string, shardNOs []uint8, db *database.SimpleLDB) *TestManager {
	agents := map[uint8]*shardAgent{}

	wsm := stateManager.NewWorldStateManager(db) // for performance consider, the world state manager is multiplexed by all agents
	sme := smartcontract.NewDemoSME()
	sme.InstallLocalUsedReadWriteFunc(wsm.ReadState, func(addr common.Address, value []byte) {
		// do nothing
		return
	})

	pools := make(map[uint8]*transactionPool)

	for i := 0; i < len(shardNOs); i++ {
		agent := newShardAgent(shardNOs[i], wsm)
		txPool := newTransactionPool(shardNOs[i])
		agents[shardNOs[i]] = agent
		pools[shardNOs[i]] = txPool
	}

	host := network.NewNodeHost("testManager", selfAddr)
	host.Start()

	return &TestManager{
		pools:             pools,
		SelfAddr:          selfAddr,
		CdntAddrs:         cdntAddrs,
		CdntPG:            network.NewPeerGroup(),
		ShardNOs:          shardNOs,
		HostClient:        host,
		worldStateManager: wsm,
		blockManager:      newBlockPost(),
		wsmMux:            sync.Mutex{},
		agents:            agents,
		savedStateMessage: make([]TestStateMsg, 0),
		localUsedSME:      sme,
	}
}

func (tm *TestManager) StartDial() {
	tasks := network.SimpleCreateDialTasks_All(tm.SelfAddr, tm.CdntAddrs, 3, false)
	tm.HostClient.Dial(tasks)
}

func (tm *TestManager) BindPeers() error {
	for _, p := range tm.CdntAddrs {
		cdntPeer, err := tm.HostClient.BindPeer(p, tm.handleCoordinatorMsg)
		if err != nil {
			return err
		}
		tm.CdntPG.AddPeer(cdntPeer)
	}
	return nil
}

func (tm *TestManager) handleCoordinatorMsg(msg network.Msg) error {
	switch msg.Code {
	case Test_EpochMsg:
		var epochMsg TestEpochCommitMsg
		if err := rlp.Decode(msg.Payload, &epochMsg); err != nil {
			return err
		}
		tm.CommitEpoch(epochMsg.Epoch)

	case Test_StateBackMsg:
		var stateMsg TestStateMsg
		if err := rlp.Decode(msg.Payload, &stateMsg); err != nil {
			return err
		}
		tm.savedStateMessage = append(tm.savedStateMessage, stateMsg)
	default:
		return fmt.Errorf("unknown message with code %d is got", msg.Code)
	}
	return nil
}

func (tm *TestManager) discordMsg(msg network.Msg) error {
	return msg.Discard()
}

func (tm *TestManager) CommitEpoch(epoch elements.Epoch) {
	tm.wsmMux.Lock()
	defer tm.wsmMux.Unlock()
	tm.worldStateManager.AppendEpoch(epoch)
	tm.worldStateManager.CommitStateSet(epoch.StateCommitSet) // the world state manager is multiplexed
	for _, agent := range tm.agents {
		agent.CommitEpoch(epoch)
	}
}

func (tm *TestManager) GenerateBlock(shardNO uint8, txs []elements.Transaction) error {
	tm.wsmMux.Lock()
	defer tm.wsmMux.Unlock()

	agent, ok := tm.agents[shardNO]
	if !ok {
		return fmt.Errorf("unknown agent with shard number %d", shardNO)
	}
	block := agent.generateBlock(txs)

	tm.blockManager.add(block)
	return nil
}

func (tm *TestManager) BlockUploadLoop() {
	uploadTicker := time.NewTicker(coes.BlockUploadTriggerInterval)
	generateTicker := time.NewTicker(coes.GenerateBlockInterval)
	for {
		select {
		case <-uploadTicker.C:
			err := tm.UploadAllBlocks()
			if err != nil {
				panic(err)
			}
		case <-generateTicker.C:
			for _, shardNO := range tm.ShardNOs {
				txns := tm.pools[shardNO].retrieveAll()
				if len(txns) == 0 {
					continue
				}
				tm.GenerateBlock(shardNO, txns)
			}

		}
	}

}

// TriggerUppConsensus triggers the coordinator with index to launch an upper
// layer consensus.
func (tm *TestManager) TriggerUppConsensus(index int) error {
	if index >= len(tm.CdntPG.Peers) {
		return fmt.Errorf("index is %d, but the max is %d", index, len(tm.CdntPG.Peers)-1)
	}
	err := tm.CdntPG.Peers[index].Send(Test_TriggerMsg, []byte{})
	if err != nil {
		return err
	}
	return nil
}

// QueryState query the coordinator with index to know its state and infos
// of last epoch.
func (tm *TestManager) QueryState(index int) error {
	if index >= len(tm.CdntPG.Peers) {
		return fmt.Errorf("index is %d, but the max is %d", index, len(tm.CdntPG.Peers)-1)
	}
	err := tm.CdntPG.Peers[index].Send(Test_StateQueryMsg, []byte{})
	if err != nil {
		return err
	}
	return nil
}

// UploadAllBlocks will retrieve all stored blocks and upload them to all
// coordinators.
func (tx *TestManager) UploadAllBlocks() error {
	blocks := tx.blockManager.retrieve()
	if len(blocks) == 0 {
		return nil
	}
	uploadBlocksMsg := TestBlocksUploadMsg{Blocks: blocks}
	return tx.CdntPG.Broadcast(Test_BlocksMsg, &uploadBlocksMsg)
}

// LoopTriggerOn will make a coordinator with given index to trigger the upper
// layer consensus round once it is at spare stage.
func (tm *TestManager) LoopTriggerOn(index int) error {
	if index >= len(tm.CdntPG.Peers) {
		return fmt.Errorf("index is %d, but the max is %d", index, len(tm.CdntPG.Peers)-1)
	}
	err := tm.CdntPG.Peers[index].Send(Test_LoopTiggerMsg, []byte{})
	if err != nil {
		return err
	}
	return nil
}

// LoopTriggerOff will stop the trigger loop of coordinator with given index
func (tm *TestManager) LoopTriggerOff(index int) error {
	if index >= len(tm.CdntPG.Peers) {
		return fmt.Errorf("index is %d, but the max is %d", index, len(tm.CdntPG.Peers)-1)
	}
	err := tm.CdntPG.Peers[index].Send(Test_StopLoopTriggerMsg, []byte{})
	if err != nil {
		return err
	}
	return nil
}

func (tm *TestManager) CurrentEpochHeight() uint64 {
	return tm.worldStateManager.CurrentHeight()
}

func (tm *TestManager) GetEpochByHeight(height uint64) (*elements.Epoch, error) {
	return tm.worldStateManager.GetEpochByHeight(height)
}

// LocalExecuteTransaction executes a transaction locally and only back result
func (tm *TestManager) LocalExecuteTransaction(tx elements.Transaction) (result []byte) {
	return tm.localUsedSME.LocalExecuteTransaction(tx)
}

func (tm *TestManager) InvokeYCSBTx(funcName string, args []string) (txID common.Hash, err error) {
	argBytes := StringsToBytes(args)
	switch funcName {
	case "read":
		if len(argBytes) != 1 {
			return common.Hash{}, fmt.Errorf("the number of args not right")
		}
		accountAddr := common.BytesToAddress(argBytes[0])
		index := ComputeRelateShardIndex(accountAddr[:], len(tm.ShardNOs))
		shardNO := tm.ShardNOs[index]
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_READ,
			Arguments:     []elements.Argument{argument1},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}
		tx.SetTxID()
		tm.pools[shardNO].addTx(tx)
		return tx.TxID, nil
	case "write":
		if len(argBytes) != 2 {
			return common.Hash{}, fmt.Errorf("the number of args not right")
		}

		accountAddr := common.BytesToAddress(argBytes[0])
		index := ComputeRelateShardIndex(accountAddr[:], len(tm.ShardNOs))
		shardNO := tm.ShardNOs[index]
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		argument2 := elements.Argument{
			Type:    0,
			Address: common.Address{},
			Value:   argBytes[1],
		}

		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_WRITE,
			Arguments:     []elements.Argument{argument1, argument2},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}
		tx.SetTxID()
		tm.pools[shardNO].addTx(tx)
		return tx.TxID, nil
	case "delete":
		if len(argBytes) != 1 {
			return common.Hash{}, fmt.Errorf("the number of args not right")
		}
		accountAddr := common.BytesToAddress(argBytes[0])
		index := ComputeRelateShardIndex(accountAddr[:], len(tm.ShardNOs))
		shardNO := tm.ShardNOs[index]
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_DELETE,
			Arguments:     []elements.Argument{argument1},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}
		tx.SetTxID()
		tm.pools[shardNO].addTx(tx)
		return tx.TxID, nil
	default:
		return common.Hash{}, fmt.Errorf("unknown function name %s", funcName)
	}
}

func (tm *TestManager) InvokeYCSBTx_Locally(funcName string, args []string) (string, error) {
	argBytes := StringsToBytes(args)
	switch funcName {
	case "read":
		if len(argBytes) != 1 {
			return "", fmt.Errorf("the number of args not right")
		}
		accountAddr := common.BytesToAddress(argBytes[0])
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_READ,
			Arguments:     []elements.Argument{argument1},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}

		res := tm.LocalExecuteTransaction(tx)
		return string(res), nil
	case "write":
		if len(argBytes) != 2 {
			return "", fmt.Errorf("the number of args not right")
		}

		accountAddr := common.BytesToAddress(argBytes[0])
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		argument2 := elements.Argument{
			Type:    0,
			Address: common.Address{},
			Value:   argBytes[1],
		}

		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_WRITE,
			Arguments:     []elements.Argument{argument1, argument2},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}

		res := tm.LocalExecuteTransaction(tx)
		return string(res), nil
	case "delete":
		if len(argBytes) != 1 {
			return "", fmt.Errorf("the number of args not right")
		}
		accountAddr := common.BytesToAddress(argBytes[0])
		argument1 := elements.Argument{
			Type:    1,
			Address: accountAddr,
			Value:   nil,
		}
		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        common.HashToAddress(common.GenerateRandomHash()),
			Version:       common.Hash{},
			Nonce:         0,
			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
			Function:      smartcontract.KVSTORE_FUNC_DELETE,
			Arguments:     []elements.Argument{argument1},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}
		res := tm.LocalExecuteTransaction(tx)
		return string(res), nil
	default:
		return "", fmt.Errorf("unknown function name %s", funcName)
	}
}
