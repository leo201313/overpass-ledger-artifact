package entity

import (
	"fmt"
	"opl/common"
	"opl/elements"
	"strings"
	"sync"
)

type w_state uint8

const (
	spare w_state = iota
	ordering_wait_transactions

	ordering_wait_prepare
	ordering_wait_prepare_leader
	ordering_wait_commit

	confirming_wait_done

	committing_wait_epochs
	epoch_update_wait_done
	epoch_update_wait_done_leader
)

type workerStates struct {
	localEpochVersion common.Hash // the viewed epoch version of the shard,and is usually lag of the global version
	newEpoch          *epochCache // the committed epoch versions received from the upper layer, usually should only contain at most 1, but if the system is tough, maybe over 1

	currentRank int // set at the beginning and view change
	rankAmount  int // the amount of workers in the shard

	stateMux    sync.Mutex
	stage       w_state
	nowHeight   uint64
	nextNonce   uint64
	currentTurn int // set at the new epoch committed. + 1 when a round of lower consensus finished

	candidateTxIDs []common.Hash
	candidateBlock *elements.Block

	ppmHolder           *lowPreprepareMsgHolder // used to store early coming preprepare message
	cacheMux            sync.RWMutex
	cache_LowPrepareMsg []LowPrepareMsg
	cache_LowCommitMsg  []LowCommitMsg
	cache_LowDoneMsg    []LowDoneMsg

	esimHolder            *lowEpochStepInMsgHolder
	cache_LowEpochDoneMsg []LowEpochDoneMsg
	candidateEpochIDs     []common.Hash

	//lus *LastUppState
}

//type LastUppState struct {
//	Used        bool
//	BaseVersion common.Hash
//	NextVersion common.Hash
//	mux         sync.Mutex
//}
//
//func NewLastUppState() *LastUppState {
//	return &LastUppState{
//		Used:        false,
//		BaseVersion: common.Hash{},
//		NextVersion: common.Hash{},
//		mux:         sync.Mutex{},
//	}
//}
//
//func (lus *LastUppState) Update(baseVersion, nextVersion common.Hash) {
//	lus.mux.Lock()
//	defer lus.mux.Unlock()
//	lus.Used = true
//	lus.BaseVersion = baseVersion
//	lus.NextVersion = nextVersion
//}
//
//func (lus *LastUppState) Retrieve() (baseVersion common.Hash, nextVersion common.Hash, used bool) {
//	lus.mux.Lock()
//	defer lus.mux.Unlock()
//	return lus.BaseVersion, lus.NextVersion, lus.Used
//}
//
//func (lus *LastUppState) Reset() {
//	lus.mux.Lock()
//	defer lus.mux.Unlock()
//	lus.Used = false
//}

func newWorkerStates(currentRank int, rankAmount int) *workerStates {
	return &workerStates{
		localEpochVersion:     common.INITIAL_VERSION,
		newEpoch:              newEpochCache(0),
		currentRank:           currentRank,
		rankAmount:            rankAmount,
		stateMux:              sync.Mutex{},
		stage:                 spare,
		nowHeight:             0,
		nextNonce:             0,
		currentTurn:           0,
		candidateTxIDs:        make([]common.Hash, 0),
		candidateBlock:        nil,
		ppmHolder:             newLowPreprepareMsgHolder(),
		cacheMux:              sync.RWMutex{},
		cache_LowPrepareMsg:   make([]LowPrepareMsg, 0),
		cache_LowCommitMsg:    make([]LowCommitMsg, 0),
		cache_LowDoneMsg:      make([]LowDoneMsg, 0),
		esimHolder:            newLowEpochStepInMsgHolder(),
		cache_LowEpochDoneMsg: make([]LowEpochDoneMsg, 0),
		candidateEpochIDs:     make([]common.Hash, 0),
		//lus:                   NewLastUppState(),
	}
}

type epochCache struct {
	nowHeight uint64
	epochs    []elements.Epoch
	mux       sync.RWMutex
}

func newEpochCache(nowHeight uint64) *epochCache {
	return &epochCache{
		nowHeight: nowHeight,
		epochs:    make([]elements.Epoch, 0),
		mux:       sync.RWMutex{},
	}
}

func (ec *epochCache) addEpoch(epoch elements.Epoch) error {
	ec.mux.Lock()
	defer ec.mux.Unlock()
	if ec.nowHeight+1 != epoch.Height {
		return fmt.Errorf("the epoch got is not in sequence, as the now heigtht is %d, want %d, but got %d", ec.nowHeight, ec.nowHeight+1, epoch.Height)
	}
	ec.epochs = append(ec.epochs, epoch)
	ec.nowHeight += 1
	return nil
}

// NOTE: epochIDs should be in sequence, or containEpoch back false
func (ec *epochCache) containEpoch(epochIDs []common.Hash) bool {
	ec.mux.RLock()
	defer ec.mux.RUnlock()

	if len(ec.epochs) < len(epochIDs) {
		return false
	}

	for i := 0; i < len(epochIDs); i++ {
		if ec.epochs[i].EpochID != epochIDs[i] {
			return false
		}
	}

	return true
}

func (ec *epochCache) isEmpty() bool {
	ec.mux.RLock()
	defer ec.mux.RUnlock()

	return len(ec.epochs) == 0
}

// usually used after isEmpty
func (ec *epochCache) selectAll() []elements.Epoch {
	ec.mux.Lock()
	defer ec.mux.Unlock()

	selectedEpochs := ec.epochs
	ec.epochs = make([]elements.Epoch, 0)
	return selectedEpochs
}

// usually used after containEpoch
func (ec *epochCache) retrieveAll(epochIDs []common.Hash) []elements.Epoch {
	ec.mux.Lock()
	defer ec.mux.Unlock()

	total := len(epochIDs)

	// Check if the number of epochIDs exceeds the cached epochs
	if total > len(ec.epochs) {
		panic("retrieveAll failed: the number of epochIDs exceeds cached epochs")
	}

	// Create a slice for retrievedEpochs and copy the first 'total' epochs
	retrievedEpochs := make([]elements.Epoch, total)
	copy(retrievedEpochs, ec.epochs[:total])

	// Validate that each epochID in epochIDs matches the corresponding EpochID in retrievedEpochs
	for i := 0; i < total; i++ {
		if epochIDs[i] != retrievedEpochs[i].EpochID {
			panic("retrieveAll failed: unmatched epochID in input")
		}
	}

	// Remove the first 'total' elements from the cache
	ec.epochs = ec.epochs[total:]

	return retrievedEpochs
}

func (ws *workerStates) processUppEpochMsg(msg UppEpochMsg) error {
	err := ws.newEpoch.addEpoch(msg.Epoch)
	return err
}

func (ws *workerStates) processLowPreprepareMsg(msg LowPreprepareMsg) error {
	ws.stateMux.Lock()
	defer ws.stateMux.Unlock()

	switch ws.stage {
	case spare: // be a follower and all done message has got
		if msg.EpochID != ws.localEpochVersion || msg.Height != ws.nowHeight {
			return fmt.Errorf("fail in processLowPreprepareMsg, expect version %x with height %d, got version %x with height %d", ws.localEpochVersion, ws.nowHeight, msg.EpochID, msg.Height)
		}
		if msg.Nonce != ws.nextNonce {
			return fmt.Errorf("fail in processLowPreprepareMsg, expect nonce %d, but got %d", ws.nextNonce, msg.Nonce)
		}

		ws.candidateTxIDs = msg.TxSequence
		ws.stage = ordering_wait_transactions
		return nil
	case confirming_wait_done: // be a follower and not all done message has got
		ws.ppmHolder.pending(msg)
		return nil
	case epoch_update_wait_done: // be a follower and the epoch update is not sure of completion
		ws.ppmHolder.pending(msg)
		return nil
	default:
		return fmt.Errorf("unexpected worker stage %d when processing LowPreprepareMsg", ws.stage)
	}
}

// only used in Worker.Operate if the ppmHolder is occupied with preprepare message
func (ws *workerStates) processLowPreprepareMsg_SP(msg LowPreprepareMsg) error {
	// the state is locked in Worker.Operate
	// and the state is known as spare.

	if msg.EpochID != ws.localEpochVersion || msg.Height != ws.nowHeight {
		return fmt.Errorf("fail in processLowPreprepareMsg_SP, expect version %x with height %d, got version %x with height %d", ws.localEpochVersion, ws.nowHeight, msg.EpochID, msg.Height)
	}
	if msg.Nonce != ws.nextNonce {
		return fmt.Errorf("fail in processLowPreprepareMsg_SP, expect nonce %d, but got %d", ws.nextNonce, msg.Nonce)
	}

	ws.candidateTxIDs = msg.TxSequence
	ws.stage = ordering_wait_transactions
	return nil
}

type lowPreprepareMsgHolder struct {
	on         bool
	pendingMsg LowPreprepareMsg
}

func (lpmh *lowPreprepareMsgHolder) pending(msg LowPreprepareMsg) {
	lpmh.pendingMsg = msg
	lpmh.on = true
}

func (lpmh *lowPreprepareMsgHolder) occupy() bool {
	return lpmh.on
}

func (lpmh *lowPreprepareMsgHolder) retrieve() LowPreprepareMsg {
	lpmh.on = false
	return lpmh.pendingMsg
}

func newLowPreprepareMsgHolder() *lowPreprepareMsgHolder {
	return &lowPreprepareMsgHolder{
		on:         false,
		pendingMsg: LowPreprepareMsg{},
	}
}

type lowEpochStepInMsgHolder struct {
	on         bool
	pendingMsg LowEpochStepInMsg
}

func (lesimh *lowEpochStepInMsgHolder) pending(msg LowEpochStepInMsg) {
	lesimh.pendingMsg = msg
	lesimh.on = true
}

func (lesimh *lowEpochStepInMsgHolder) occupy() bool {
	return lesimh.on
}

func (lesimh *lowEpochStepInMsgHolder) retrieve() LowEpochStepInMsg {
	lesimh.on = false
	return lesimh.pendingMsg
}

func newLowEpochStepInMsgHolder() *lowEpochStepInMsgHolder {
	return &lowEpochStepInMsgHolder{
		on:         false,
		pendingMsg: LowEpochStepInMsg{},
	}
}

func (ws *workerStates) addLowPrepareMsg(msg LowPrepareMsg) {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()

	ws.cache_LowPrepareMsg = append(ws.cache_LowPrepareMsg, msg)
}

func (ws *workerStates) addLowCommitMsg(msg LowCommitMsg) {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()

	ws.cache_LowCommitMsg = append(ws.cache_LowCommitMsg, msg)
}

func (ws *workerStates) addLowDoneMsg(msg LowDoneMsg) {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()

	ws.cache_LowDoneMsg = append(ws.cache_LowDoneMsg, msg)
}

func (ws *workerStates) addLowEpochDoneMsg(msg LowEpochDoneMsg) {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()

	ws.cache_LowEpochDoneMsg = append(ws.cache_LowEpochDoneMsg, msg)
}

func (ws *workerStates) IsPrepare(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1, preprepare msg -1, so here should +2
	if len(ws.cache_LowPrepareMsg)+2 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowPrepareMsg {
			if msg.Nonce == ws.nextNonce && msg.Height == ws.nowHeight {
				count += 1
			}
		}
		if count+2 >= threshold {
			return true
		} else {
			return false
		}
	}
}

// IsPrepare_Leader is only used for the leader
func (ws *workerStates) IsPrepare_Leader(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1, so here should +1
	if len(ws.cache_LowPrepareMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowPrepareMsg {
			if msg.Nonce == ws.nextNonce && msg.Height == ws.nowHeight {
				count += 1
			}
		}
		if count+1 >= threshold {
			return true
		} else {
			return false
		}
	}
}

func (ws *workerStates) IsCommit(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1
	if len(ws.cache_LowCommitMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowCommitMsg {
			if msg.Nonce == ws.nextNonce && msg.Height == ws.nowHeight {
				count += 1
			}
		}
		if count+1 >= threshold {
			return true
		} else {
			return false
		}
	}
}

func (ws *workerStates) IsConfirmed(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1
	if len(ws.cache_LowDoneMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowDoneMsg {
			if msg.Nonce == ws.nextNonce && msg.Height == ws.nowHeight {
				count += 1
			}
		}
		if count+1 >= threshold {
			return true
		} else {
			return false
		}
	}
}

func (ws *workerStates) IsEpochUpdateDone(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1, step in message -1, so here should +2
	if len(ws.cache_LowEpochDoneMsg)+2 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowEpochDoneMsg {
			if msg.NowHeight == ws.nowHeight {
				count += 1
			}
		}
		if count+2 >= threshold {
			return true
		} else {
			return false
		}
	}
}

// IsEpochUpdateDone_Leader only used for leader of the epoch commitment
func (ws *workerStates) IsEpochUpdateDone_Leader(threshold int) bool {
	ws.cacheMux.RLock()
	defer ws.cacheMux.RUnlock()

	// self -1, so here should +1
	if len(ws.cache_LowEpochDoneMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range ws.cache_LowEpochDoneMsg {
			if msg.NowHeight == ws.nowHeight {
				count += 1
			}
		}
		if count+1 >= threshold {
			return true
		} else {
			return false
		}
	}
}

func (ws *workerStates) StepNextTurn() {
	ws.nextNonce += 1
	ws.currentTurn = (ws.currentTurn + 1) % ws.rankAmount
}

func (ws *workerStates) EmptyCache(targetHeight uint64, targetNonce uint64) {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()
	cache_LowPrepareMsg_ := make([]LowPrepareMsg, 0)
	cache_LowCommitMsg_ := make([]LowCommitMsg, 0)
	cache_LowDoneMsg_ := make([]LowDoneMsg, 0)

	for _, msg := range ws.cache_LowPrepareMsg {
		if msg.Height == targetHeight && msg.Nonce == targetNonce {
			continue
		}
		cache_LowPrepareMsg_ = append(cache_LowPrepareMsg_, msg)
	}

	for _, msg := range ws.cache_LowCommitMsg {
		if msg.Height == targetHeight && msg.Nonce == targetNonce {
			continue
		}
		cache_LowCommitMsg_ = append(cache_LowCommitMsg_, msg)
	}

	for _, msg := range ws.cache_LowDoneMsg {
		if msg.Height == targetHeight && msg.Nonce == targetNonce {
			continue
		}
		cache_LowDoneMsg_ = append(cache_LowDoneMsg_, msg)
	}

	ws.cache_LowPrepareMsg = cache_LowPrepareMsg_
	ws.cache_LowCommitMsg = cache_LowCommitMsg_
	ws.cache_LowDoneMsg = cache_LowDoneMsg_
}

func (ws *workerStates) EmptyEpochDoneMsg() {
	ws.cacheMux.Lock()
	defer ws.cacheMux.Unlock()
	cache_LowEpochDoneMsg := make([]LowEpochDoneMsg, 0)
	ws.cache_LowEpochDoneMsg = cache_LowEpochDoneMsg
}

func (ws *workerStates) processEpochStepInMsg(msg LowEpochStepInMsg) error {
	ws.stateMux.Lock()
	defer ws.stateMux.Unlock()
	switch ws.stage {
	case spare:
		if msg.PrevHeight != ws.nowHeight {
			return fmt.Errorf("fail in processEpochStepInMsg, now height is %d, the message's previous height is %d", ws.nowHeight, msg.PrevHeight)
		}
		ws.candidateEpochIDs = msg.EpochIDs
		ws.stage = committing_wait_epochs
		return nil
	case confirming_wait_done:
		ws.esimHolder.pending(msg)
		return nil
	default:
		fmt.Errorf("unexpected worker stage %d when processing LowEpochStepInMsg", ws.stage)
	}
	return nil
}

// only used in Worker.Operate if the esimHolder is occupied with epoch step in message
func (ws *workerStates) processEpochStepInMsg_SP(msg LowEpochStepInMsg) error {
	// the state is locked in Worker.Operate
	// and the state is known as spare.
	if msg.PrevHeight != ws.nowHeight {
		return fmt.Errorf("fail in processEpochStepInMsg, now height is %d, the message's previous height is %d", ws.nowHeight, msg.PrevHeight)
	}
	ws.candidateEpochIDs = msg.EpochIDs
	ws.stage = committing_wait_epochs
	return nil
}

//---------------------------------String functions

//// String method for LastUppState
//// Returns a formatted string representation of the LastUppState struct.
//func (l LastUppState) String() string {
//	return fmt.Sprintf("Used: %v, BaseVersion: %x, NextVersion: %x", l.Used, l.BaseVersion, l.NextVersion)
//}

// String method for epochCache
// Returns a formatted string representation of the epochCache struct.
func (e epochCache) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Now Height: %d, Epochs Count: %d\n", e.nowHeight, len(e.epochs)))
	// 如果 epochs 有内容，可以进一步打印
	if len(e.epochs) > 0 {
		builder.WriteString("Containing Epochs: ")
		for _, epoch := range e.epochs {
			builder.WriteString(fmt.Sprintf("EpochID: %x", epoch.EpochID))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

// String method for lowPreprepareMsgHolder
// Returns a formatted string representation of the lowPreprepareMsgHolder struct.
func (l *lowPreprepareMsgHolder) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("the On is: %v\n", l.on))
	if l.on {
		builder.WriteString(fmt.Sprintf("The pending LowPrepareMessage is:\n %s\n", l.pendingMsg.String()))
	}
	return builder.String()
}

// String method for lowEpochStepInMsgHolder
// Returns a formatted string representation of the lowEpochStepInMsgHolder struct.
func (l *lowEpochStepInMsgHolder) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("the On is: %v\n", l.on))
	if l.on {
		builder.WriteString(fmt.Sprintf("The pending LowEpochStepInMsg is:\n %s\n", l.pendingMsg.String()))
	}
	return builder.String()
}

// String method for workerStates
// Returns a formatted string representation of the workerStates struct.
func (ws workerStates) String() string {
	var builder strings.Builder

	// Local epoch version
	builder.WriteString(fmt.Sprintf("Local Epoch Version: %x\n", ws.localEpochVersion))

	// New epoch information, if available
	if ws.newEpoch != nil {
		builder.WriteString(fmt.Sprintf("New Epoch: %s\n", ws.newEpoch.String()))
	} else {
		builder.WriteString("New Epoch: nil\n")
	}

	// Worker rank information
	builder.WriteString(fmt.Sprintf("Current Rank: %d\n", ws.currentRank))
	builder.WriteString(fmt.Sprintf("Rank Amount: %d\n", ws.rankAmount))

	// Worker state (stage)
	builder.WriteString(fmt.Sprintf("State: %v\n", ws.stage))

	// Current height, nonce, and turn
	builder.WriteString(fmt.Sprintf("Current Height: %d\n", ws.nowHeight))
	builder.WriteString(fmt.Sprintf("Next Nonce: %d\n", ws.nextNonce))
	builder.WriteString(fmt.Sprintf("Current Turn: %d\n", ws.currentTurn))

	// Candidate transactions and block information
	builder.WriteString(fmt.Sprintf("Candidate TxIDs: %d\n", len(ws.candidateTxIDs)))
	if ws.candidateBlock != nil {
		builder.WriteString(fmt.Sprintf("Candidate Block: %x\n", ws.candidateBlock.BlockID))
	}

	// Pre-prepare message holder and cache information
	if ws.ppmHolder != nil {
		builder.WriteString(fmt.Sprintf("PPM Holder: %s\n", ws.ppmHolder.String()))
	} else {
		builder.WriteString("PPM Holder: nil\n")
	}
	builder.WriteString(fmt.Sprintf("Cache LowPrepareMsg Count: %d\n", len(ws.cache_LowPrepareMsg)))
	builder.WriteString(fmt.Sprintf("Cache LowCommitMsg Count: %d\n", len(ws.cache_LowCommitMsg)))
	builder.WriteString(fmt.Sprintf("Cache LowDoneMsg Count: %d\n", len(ws.cache_LowDoneMsg)))

	// Epoch step-in message holder and cache information
	if ws.esimHolder != nil {
		builder.WriteString(fmt.Sprintf("ESIM Holder: %s\n", ws.esimHolder.String()))
	} else {
		builder.WriteString("ESIM Holder: nil\n")
	}
	builder.WriteString(fmt.Sprintf("Cache LowEpochDoneMsg Count: %d\n", len(ws.cache_LowEpochDoneMsg)))

	// Candidate epoch IDs
	builder.WriteString(fmt.Sprintf("Candidate Epoch IDs: %d\n", len(ws.candidateEpochIDs)))

	//// Last upper state information, if available
	//if ws.lus != nil {
	//	builder.WriteString(fmt.Sprintf("Last Upper State: %s\n", ws.lus.String()))
	//} else {
	//	builder.WriteString("Last Upper State: nil\n")
	//}

	return builder.String()
}
