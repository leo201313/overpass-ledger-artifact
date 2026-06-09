package entity

import (
	"fmt"
	"github.com/golang/glog"
	"opl/common"
	"opl/elements"
	"opl/logger"
	"strings"
	"sync"
	"time"
)

type blockCache struct {
	channels map[uint8]*blockChannel

	newWaitBegin  bool
	waitStartTime time.Time
}

// blockChannel is the channel to receive block from its connected shard
type blockChannel struct {
	shardNO      uint8
	nowPointer   *pointer
	storedBlocks []elements.Block
	accessMux    sync.RWMutex
}

type pointer struct {
	currentVersion common.Hash
	index          uint64

	previousVersion common.Hash // at this version, the max lag is 2
}

// this tailored_version is just used for BlockChannel at the initialization
var tailored_version = common.BytesToHash([]byte("tailored_version_channel"))

func newBlockChannel(shardNO uint8) *blockChannel {
	return &blockChannel{
		shardNO: shardNO,
		nowPointer: &pointer{
			currentVersion:  tailored_version,
			index:           0,
			previousVersion: tailored_version,
		},
		storedBlocks: make([]elements.Block, 0),
	}
}

func (bc *blockChannel) add(block elements.Block) error {
	bc.accessMux.Lock()
	defer bc.accessMux.Unlock()

	if block.Version == bc.nowPointer.currentVersion {
		if block.Nonce != (bc.nowPointer.index + 1) {
			return fmt.Errorf("the block added is not in sequence")
		} else {
			bc.nowPointer.index += 1
			bc.storedBlocks = append(bc.storedBlocks, block)
			return nil
		}
	} else { // the block maybe based on a new version
		if block.Version == bc.nowPointer.previousVersion {
			return fmt.Errorf("the block added is based on a previous version")
		} else {
			if block.Nonce != 0 {
				return fmt.Errorf("a new version is found but the initial sequence is not 0")
			}
			bc.nowPointer.previousVersion = bc.nowPointer.currentVersion
			bc.nowPointer.currentVersion = block.Version
			bc.nowPointer.index = 0
			bc.storedBlocks = append(bc.storedBlocks, block)
			return nil
		}
	}
}

// the target Blocks must input with the same sequence
func (bc *blockChannel) findAll(targetBlocks []common.Hash) bool {
	bc.accessMux.RLock()
	defer bc.accessMux.RUnlock()

	//outer:
	//	for _, id := range targetBlocks {
	//		for _, blk := range bc.storedBlocks {
	//			if id == blk.BlockID {
	//				continue outer
	//			}
	//		}
	//		return false
	//	}
	//
	//	return true

	if len(bc.storedBlocks) < len(targetBlocks) {
		return false
	}

	for i := 0; i < len(targetBlocks); i++ {
		if targetBlocks[i] != bc.storedBlocks[i].BlockID {
			glog.V(logger.Error).Infof("unexpected block sequence is found in blockChannel with block %x and block %x", targetBlocks[i], bc.storedBlocks[i].BlockID)
			panic("unexpected block sequence is found")
			return false
		}
	}

	return true
}

// retrieveAll must be used after known all blockIDs are found in the channel
func (bc *blockChannel) retrieveAll(targetBlocks []common.Hash) []elements.Block {
	bc.accessMux.Lock()
	defer bc.accessMux.Unlock()

	count := len(targetBlocks)
	retrieved := make([]elements.Block, count)
	copy(retrieved, bc.storedBlocks[:count])
	bc.storedBlocks = bc.storedBlocks[count:]
	return retrieved
}

// readAll is only used for test
func (bc *blockChannel) readAll() []elements.Block {
	bc.accessMux.RLock()
	defer bc.accessMux.RUnlock()

	readBlocks := make([]elements.Block, len(bc.storedBlocks))
	copy(readBlocks, bc.storedBlocks)
	return readBlocks
}

// selectAll selects all stored blocks in the channel
func (bc *blockChannel) selectAll() []common.Hash {
	bc.accessMux.RLock()
	defer bc.accessMux.RUnlock()

	blockIDs := make([]common.Hash, len(bc.storedBlocks))

	for i, block := range bc.storedBlocks {
		blockIDs[i] = block.BlockID
	}
	return blockIDs
}

// selectAll_SP selects all stored blocks in the channel with additional info
func (bc *blockChannel) selectAll_SP() ([]common.Hash, []common.Hash, []uint64) {
	bc.accessMux.RLock()
	defer bc.accessMux.RUnlock()

	blockIDs := make([]common.Hash, len(bc.storedBlocks))
	baseVersions := make([]common.Hash, len(bc.storedBlocks))
	nonces := make([]uint64, len(bc.storedBlocks))
	for i, block := range bc.storedBlocks {
		blockIDs[i] = block.BlockID
		baseVersions[i] = block.Version
		nonces[i] = block.Nonce
	}
	return blockIDs, baseVersions, nonces
}

func newBlockCache(shardNOs []uint8) *blockCache {
	channels := make(map[uint8]*blockChannel)
	for _, shardNO := range shardNOs {
		channel := newBlockChannel(shardNO)
		channels[shardNO] = channel
	}
	return &blockCache{channels: channels, newWaitBegin: false}
}

func (bk *blockCache) add(shardNO uint8, block elements.Block) error {
	if channel, ok := bk.channels[shardNO]; !ok {
		return fmt.Errorf("the shard %d is not found", shardNO)
	} else {
		return channel.add(block)
	}
}

// addBlock is just used for test
func (bk *blockCache) addBlock(block elements.Block) error {
	if channel, ok := bk.channels[block.ShardNO]; !ok {
		return fmt.Errorf("the shard %d is not found", block.ShardNO)
	} else {
		return channel.add(block)
	}
}

func (bk *blockCache) findAll(targetBlocks map[uint8][]common.Hash) bool {
	for shardNO, blockIDs := range targetBlocks {
		if channel, ok := bk.channels[shardNO]; !ok {
			glog.V(logger.Error).Infof("unknown shard number is found: %d", shardNO)
			panic("fail in blockCache.findAll")
		} else {
			if channel.findAll(blockIDs) {
				continue
			} else {
				return false
			}
		}
	}

	return true
}

func (bk *blockCache) retrieveAll(targetBlocks map[uint8][]common.Hash) map[uint8][]elements.Block {
	retrieved := make(map[uint8][]elements.Block)
	for shardNO, blockIDs := range targetBlocks {
		if channel, ok := bk.channels[shardNO]; !ok {
			glog.V(logger.Error).Infof("unknown shard number is found: %d", shardNO)
			panic("fail in blockCache.retrieveAll")
		} else {
			retrieved[shardNO] = channel.retrieveAll(blockIDs)
		}
	}
	return retrieved
}

// only used for test
func (bk *blockCache) readAll() map[uint8][]elements.Block {
	allBlocks := make(map[uint8][]elements.Block)
	for shardNO, channel := range bk.channels {
		allBlocks[shardNO] = channel.readAll()
	}
	return allBlocks
}

//// Used for specific condition
//// selectStrict will select only one block form different channels one by one
//func (bk *blockCache) selectStrict() (map[uint8][]common.Hash, bool) {
//	allBlocks := make(map[uint8][]common.Hash)
//
//	gotAll := true
//
//	for shardNO, channel := range bk.channels {
//		IDlst, got := channel.selectStrict()
//		if got {
//			bk.updateWaitRelated()
//		} else {
//			gotAll = false
//		}
//		allBlocks[shardNO] = IDlst
//	}
//
//	if gotAll {
//		bk.resetWaitRelated()
//		return allBlocks, true
//	} else {
//		if bk.newWaitBegin {
//			delta := time.Now().Sub(bk.waitStartTime)
//			if delta > coes.StrictMaxWaitDelay {
//				// has waited for a long time, should back true
//				bk.resetWaitRelated()
//				return allBlocks, true
//			} else {
//				return nil, false
//			}
//		} else {
//			return nil, false
//		}
//	}
//
//	return nil, false
//}

//func (bk *blockCache) updateWaitRelated() {
//	if bk.newWaitBegin {
//		return
//	} else {
//		bk.newWaitBegin = true
//		bk.waitStartTime = time.Now()
//	}
//}
//
//func (bk *blockCache) resetWaitRelated() {
//	bk.newWaitBegin = false
//}

// selectAll_SP selects all stored blocks in all channels, and
// decides whether to launch the new consensus round based on specific conditions.
func (bk *blockCache) selectAll_SP() (map[uint8][]common.Hash, bool) {
	// Variable to track if all conditions are met for launching the consensus round
	baseVersion := common.INITIAL_VERSION
	allBlockIDs := make(map[uint8][]common.Hash)

	// Iterate over each shard and its associated channel
	for shardNO, channel := range bk.channels {
		// Get block IDs, versions, and nonces for the channel
		ids, versions, nonces := channel.selectAll_SP()

		// If the channel has no blocks, skip it
		if len(ids) == 0 {
			return nil, false
		}

		if baseVersion == common.INITIAL_VERSION {
			baseVersion = versions[0]
		}

		if versions[0] != baseVersion {
			panic(fmt.Errorf("The block %x based on version %x with nonce %d is different from others with base version %x", ids[0], versions[0], nonces[0], baseVersion))
			return nil, false
		}

		allBlockIDs[shardNO] = ids[0:1]
	}

	return allBlockIDs, true
}

// selectAll selects all blocks stored in all channels
func (bk *blockCache) selectAll() map[uint8][]common.Hash {
	allBlocks := make(map[uint8][]common.Hash)
	for shardNO, channel := range bk.channels {
		allBlocks[shardNO] = channel.selectAll()
	}
	return allBlocks
}

func SeparateSequence(blockSequence []common.Hash, blockAssociated []uint8) (map[uint8][]common.Hash, error) {
	blockIDs := make(map[uint8][]common.Hash)
	if len(blockSequence) != len(blockAssociated) {
		return blockIDs, fmt.Errorf("the number of blockSequence not match the associated shards")
	}
	for i := 0; i < len(blockSequence); i++ {
		if _, ok := blockIDs[blockAssociated[i]]; ok {
			blockIDs[blockAssociated[i]] = append(blockIDs[blockAssociated[i]], blockSequence[i])
		} else {
			initialGroup := make([]common.Hash, 0)
			initialGroup = append(initialGroup, blockSequence[i])
			blockIDs[blockAssociated[i]] = initialGroup
		}
	}
	return blockIDs, nil
}

func AggregateSequence(groupedBlocks map[uint8][]elements.Block, blockSequence []common.Hash, blockAssociated []uint8) ([]elements.Block, error) {
	// Create a result slice to store the aggregated blocks
	result := make([]elements.Block, 0, len(blockSequence))

	// Check if the input lengths match
	if len(blockSequence) != len(blockAssociated) {
		return nil, fmt.Errorf("the number of blockSequence not match the associated shards")
	}

	// Track the current index for each shard
	currentIndex := make(map[uint8]int)

	// Iterate over blockSequence and blockAssociated
	for i := 0; i < len(blockSequence); i++ {
		shardID := blockAssociated[i]
		index := currentIndex[shardID]

		// Check if the shard exists and has enough blocks
		if index >= len(groupedBlocks[shardID]) {
			return nil, fmt.Errorf("not enough blocks in shard %d", shardID)
		}

		tempBlock := groupedBlocks[shardID][index]
		if tempBlock.BlockID != blockSequence[i] {
			return nil, fmt.Errorf("the blockID is not match, with in sequence: %x, in blocks: %x", blockSequence[i], tempBlock.BlockID)
		}

		// Append the block from groupedBlocks to the result
		result = append(result, tempBlock)
		// Update the current index for the shard
		currentIndex[shardID]++
	}

	return result, nil
}

//----------------------------String functions

// String method for blockCache
// Returns a formatted string representation of the blockCache struct.
func (b *blockCache) String() string {
	var builder strings.Builder

	// Lock for reading the channels
	builder.WriteString(fmt.Sprintf("blockCache - New Wait Begin: %v\n", b.newWaitBegin))
	builder.WriteString(fmt.Sprintf("Wait Start Time: %v\n", b.waitStartTime))

	// Print the channels in the blockCache
	builder.WriteString("Channels:\n")
	for shardNO, channel := range b.channels {
		builder.WriteString(fmt.Sprintf("  Shard %d: %s\n", shardNO, channel.String()))
	}

	return builder.String()
}

// String method for blockChannel
// Returns a formatted string representation of the blockChannel struct.
func (bc *blockChannel) String() string {
	var builder strings.Builder

	// Print the shard number
	builder.WriteString(fmt.Sprintf("ShardNO: %d\n", bc.shardNO))

	// Print the current pointer (assuming *pointer has a String method)
	builder.WriteString(fmt.Sprintf("Current Pointer: %s\n", bc.nowPointer.String()))

	// Print the number of blocks stored
	builder.WriteString(fmt.Sprintf("Stored Blocks: %d\n", len(bc.storedBlocks)))

	// Optionally, print details of the stored blocks (assuming elements.Block has a String method)
	if len(bc.storedBlocks) > 0 {
		builder.WriteString("Stored Block Details:\n")
		for i, block := range bc.storedBlocks {
			builder.WriteString(fmt.Sprintf("  Block %d: \n", i))
			builder.WriteString(fmt.Sprintf("  - BlockID     : %x\n", block.BlockID))
			builder.WriteString(fmt.Sprintf("  - BaseVersion : %x\n", block.Version))
			builder.WriteString(fmt.Sprintf("  - Nonce       : %d\n", block.Nonce))
			builder.WriteString(fmt.Sprintf("  - Transactions: %d\n", len(block.Transactions)))
		}
	} else {
		builder.WriteString("No blocks stored.\n")
	}

	return builder.String()
}

// String method for pointer
// Returns a formatted string representation of the pointer struct.
func (p *pointer) String() string {
	var builder strings.Builder

	// Print current version and index
	builder.WriteString(fmt.Sprintf("Current Version: %x\n", p.currentVersion))
	builder.WriteString(fmt.Sprintf("Index: %d\n", p.index))

	// Print previous version if available
	builder.WriteString(fmt.Sprintf("Previous Version: %x\n", p.previousVersion))

	return builder.String()
}
