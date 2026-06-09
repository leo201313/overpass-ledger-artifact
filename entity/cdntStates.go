package entity

import (
	"fmt"
	"opl/common"
	"opl/elements"
	"strings"
	"sync"
)

type c_state uint8

const (
	stage_spare c_state = iota
	stage_ordering_wait_blocks
	stage_ordering_wait_prepare

	stage_ordering_wait_prepare_leader

	stage_ordering_wait_commit
	stage_confirming_wait

	stage_confiming_wait_but_new_round_begin
)

type cdntState struct {
	nowVersion  common.Hash // current committed epoch version
	nowHeight   uint64      // the height of current version
	nextVersion common.Hash // the candidate epoch version

	stage    c_state
	stageMux sync.Mutex

	selfRank    int
	rankAmount  int
	currentTurn int

	candidateBlockSequence   []common.Hash
	candidateShardAssociated []uint8
	candidateMap             map[uint8][]common.Hash
	candidateBlocks          []elements.Block
	nowEpoch                 *elements.Epoch

	cacheMux            sync.RWMutex
	cache_UppPrepareMsg []UppPrepareMsg
	cache_UppCommitMsg  []UppCommitMsg
	cache_UppDoneMsg    []UppDoneMsg
}

func newCdntState(selfRank int, rankAmount int) *cdntState {
	return &cdntState{
		nowVersion:               common.INITIAL_VERSION,
		nowHeight:                0,
		nextVersion:              common.Hash{},
		stage:                    stage_spare,
		stageMux:                 sync.Mutex{},
		selfRank:                 selfRank,
		rankAmount:               rankAmount,
		currentTurn:              0,
		candidateBlockSequence:   nil,
		candidateShardAssociated: nil,
		candidateMap:             nil,
		candidateBlocks:          nil,
		cacheMux:                 sync.RWMutex{},
		cache_UppPrepareMsg:      make([]UppPrepareMsg, 0),
		cache_UppCommitMsg:       make([]UppCommitMsg, 0),
		cache_UppDoneMsg:         make([]UppDoneMsg, 0),
	}
}

func (cs *cdntState) addUppPrepareMsg(msg UppPrepareMsg) {
	cs.cacheMux.Lock()
	defer cs.cacheMux.Unlock()

	cs.cache_UppPrepareMsg = append(cs.cache_UppPrepareMsg, msg)
}

func (cs *cdntState) addUppCommitMsg(msg UppCommitMsg) {
	cs.cacheMux.Lock()
	defer cs.cacheMux.Unlock()

	cs.cache_UppCommitMsg = append(cs.cache_UppCommitMsg, msg)
}

func (cs *cdntState) addUppDoneMsg(msg UppDoneMsg) {
	cs.cacheMux.Lock()
	defer cs.cacheMux.Unlock()

	cs.cache_UppDoneMsg = append(cs.cache_UppDoneMsg, msg)
}

func (cs *cdntState) processUppPreprepareMsg(msg UppPreprepareMsg) error {
	cs.stageMux.Lock()
	defer cs.stageMux.Unlock()

	switch cs.stage {
	case stage_spare: // standard condition
		if msg.Height != cs.nowHeight+1 {
			return fmt.Errorf("fail in processUppPreprepareMsg, height %d is not match %d + 1", msg.Height, cs.nowHeight)
		}

		cs.nextVersion = msg.EpochID
		cs.candidateBlockSequence = msg.BlockSequence
		cs.candidateShardAssociated = msg.BlockAssociated
		targetMap, err := SeparateSequence(cs.candidateBlockSequence, cs.candidateShardAssociated)
		if err != nil {
			panic(err)
		}
		cs.candidateMap = targetMap

		cs.stage = stage_ordering_wait_blocks
		return nil
	case stage_confirming_wait: // the previous epoch is not committed but new round is coming

		if msg.Height != cs.nowHeight+2 {
			return fmt.Errorf("fail in processUppPreprepareMsg, height %d is not match %d + 2", msg.Height, cs.nowHeight)
		}

		cs.nowVersion = cs.nextVersion
		cs.nowHeight += 1

		cs.nextVersion = msg.EpochID
		cs.candidateBlockSequence = msg.BlockSequence
		cs.candidateShardAssociated = msg.BlockAssociated
		cs.currentTurn = (cs.currentTurn + 1) % cs.rankAmount

		targetMap, err := SeparateSequence(cs.candidateBlockSequence, cs.candidateShardAssociated)
		if err != nil {
			panic(err)
		}
		cs.candidateMap = targetMap

		cs.stage = stage_confiming_wait_but_new_round_begin
		return nil
	default:
		return fmt.Errorf("unexpected coordinator stage %d when processing UppPreprepareMsg", cs.stage)
	}
}

func (cs *cdntState) IsPrepare(threshold int) bool {
	cs.cacheMux.RLock()
	defer cs.cacheMux.RUnlock()

	// self -1, preprepare msg -1, so here should +2
	if len(cs.cache_UppPrepareMsg)+2 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range cs.cache_UppPrepareMsg {
			if msg.EpochID == cs.nextVersion {
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
func (cs *cdntState) IsPrepare_Leader(threshold int) bool {
	cs.cacheMux.RLock()
	defer cs.cacheMux.RUnlock()

	// self -1, so here should +1
	if len(cs.cache_UppPrepareMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range cs.cache_UppPrepareMsg {
			if msg.EpochID == cs.nextVersion {
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

func (cs *cdntState) IsCommit(threshold int) bool {
	cs.cacheMux.RLock()
	defer cs.cacheMux.RUnlock()

	// self -1
	if len(cs.cache_UppCommitMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range cs.cache_UppCommitMsg {
			if msg.EpochID == cs.nextVersion {
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

func (cs *cdntState) IsConfirmed(threshold int) bool {
	cs.cacheMux.RLock()
	defer cs.cacheMux.RUnlock()

	// self -1
	if len(cs.cache_UppDoneMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range cs.cache_UppDoneMsg {
			if msg.EpochID == cs.nextVersion {
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

func (cs *cdntState) IsConfirmed_Variant(threshold int) bool {
	cs.cacheMux.RLock()
	defer cs.cacheMux.RUnlock()

	// self -1
	if len(cs.cache_UppDoneMsg)+1 < threshold {
		return false
	} else {
		count := 0
		for _, msg := range cs.cache_UppDoneMsg {
			if msg.EpochID == cs.nowVersion {
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

func (cs *cdntState) EmptyCaches(targetVersion common.Hash) {
	cs.cacheMux.Lock()
	defer cs.cacheMux.Unlock()

	cache_UppPrepareMsg_ := make([]UppPrepareMsg, 0)
	cache_UppCommitMsg_ := make([]UppCommitMsg, 0)
	cache_UppDoneMsg_ := make([]UppDoneMsg, 0)

	for _, msg := range cs.cache_UppPrepareMsg {
		if msg.EpochID == targetVersion {
			continue
		}
		cache_UppPrepareMsg_ = append(cache_UppPrepareMsg_, msg)
	}

	for _, msg := range cs.cache_UppCommitMsg {
		if msg.EpochID == targetVersion {
			continue
		}
		cache_UppCommitMsg_ = append(cache_UppCommitMsg_, msg)
	}

	for _, msg := range cs.cache_UppDoneMsg {
		if msg.EpochID == targetVersion {
			continue
		}
		cache_UppDoneMsg_ = append(cache_UppDoneMsg_, msg)
	}

	cs.cache_UppPrepareMsg = cache_UppPrepareMsg_
	cs.cache_UppCommitMsg = cache_UppCommitMsg_
	cs.cache_UppDoneMsg = cache_UppDoneMsg_
}

// used when stage is at stage_confirming_wait after confirming
// the stageMux should be locked before calling Complete
func (cs *cdntState) Complete() {
	cs.nowVersion = cs.nextVersion
	cs.nowHeight += 1

	cs.candidateBlockSequence = nil
	cs.candidateShardAssociated = nil
	cs.candidateMap = nil
	cs.candidateBlocks = nil
	cs.nowEpoch = nil

	cs.currentTurn = (cs.currentTurn + 1) % cs.rankAmount

	cs.EmptyCaches(cs.nowVersion)
}

// used when stage is at stage_confiming_wait_but_new_round_begin after confirming
// the stageMux should be locked before calling Complete_Variant
func (cs *cdntState) Complete_Variant() {
	cs.EmptyCaches(cs.nowVersion)
}

// String方法实现
func (s cdntState) String() string {
	// 生成字段信息的字符串
	var builder strings.Builder

	// 当前版本信息
	builder.WriteString(fmt.Sprintf("Current Version: %x\n", s.nowVersion))
	builder.WriteString(fmt.Sprintf("Current Height: %d\n", s.nowHeight))

	// 下一版本信息
	builder.WriteString(fmt.Sprintf("Next Version: %x\n", s.nextVersion))

	// 阶段信息
	builder.WriteString(fmt.Sprintf("Stage: %v\n", s.stage))

	// 排名信息
	builder.WriteString(fmt.Sprintf("Self Rank: %d\n", s.selfRank))
	builder.WriteString(fmt.Sprintf("Rank Amount: %d\n", s.rankAmount))
	builder.WriteString(fmt.Sprintf("Current Turn: %d\n", s.currentTurn))

	// 候选区块信息
	builder.WriteString(fmt.Sprintf("Candidate Block Sequence: %d\n", len(s.candidateBlockSequence)))
	builder.WriteString(fmt.Sprintf("Candidate Shard Associated: %v\n", s.candidateShardAssociated))

	// 候选区块映射信息
	builder.WriteString("Candidate Map:\n")
	for shard, blocks := range s.candidateMap {
		builder.WriteString(fmt.Sprintf("  Shard %d: %d\n", shard, len(blocks)))
	}

	// 当前纪元信息
	if s.nowEpoch != nil {
		builder.WriteString(fmt.Sprintf("Current Epoch: %x\n", s.nowEpoch.EpochID))
	}

	// 缓存信息
	builder.WriteString(fmt.Sprintf("Cache UppPrepareMsg: %d\n", len(s.cache_UppPrepareMsg)))
	builder.WriteString(fmt.Sprintf("Cache UppCommitMsg: %d\n", len(s.cache_UppCommitMsg)))
	builder.WriteString(fmt.Sprintf("Cache UppDoneMsg: %d\n", len(s.cache_UppDoneMsg)))

	return builder.String()
}
