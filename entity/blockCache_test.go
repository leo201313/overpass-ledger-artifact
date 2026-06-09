package entity

import (
	"opl/common"
	"opl/elements"
	"testing"
)

var (
	epochVersion_0 = common.INITIAL_VERSION
	epochVersion_1 = common.BytesToHash([]byte("epoch1"))
	epochVersion_2 = common.BytesToHash([]byte("epoch2"))
	epochVersion_3 = common.BytesToHash([]byte("epoch3"))
	epochVersion_4 = common.BytesToHash([]byte("epoch4"))

	epochVersions = []common.Hash{epochVersion_0, epochVersion_1, epochVersion_2, epochVersion_3, epochVersion_4}
)

func generateBlocks(shardNO uint8) []elements.Block {
	blocks := make([]elements.Block, 0)
	for _, version := range epochVersions {
		for i := 0; i < 10; i++ {
			block := elements.Block{
				BlockID:      common.Hash{},
				ShardNO:      shardNO,
				Version:      version,
				Nonce:        uint64(i),
				Transactions: nil,
			}
			block.SetBlockID()
			blocks = append(blocks, block)
		}
	}
	return blocks
}

var shardNumbers_test = []uint8{0, 1, 2, 3, 4, 5}

func TestBlockCache(t *testing.T) {
	bk := newBlockCache(shardNumbers_test)
	shardBlocks := make(map[uint8][]elements.Block)
	blockIDs := make(map[uint8][]common.Hash)
	for _, shardNO := range shardNumbers_test {
		blocks := generateBlocks(shardNO)
		shardBlocks[shardNO] = blocks
		tempIDs := make([]common.Hash, 0)
		for i := 0; i < len(blocks); i++ {
			tempIDs = append(tempIDs, blocks[i].BlockID)
		}
		blockIDs[shardNO] = tempIDs
	}

	for i := 0; i < 25; i++ {
		for shardNO, group := range shardBlocks {
			err := bk.add(shardNO, group[i])
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if bk.findAll(blockIDs) {
		t.Fatal("Wrong")
	}

	halfBlocks := bk.selectAll()

	for i := 25; i < 50; i++ {
		for shardNO, group := range shardBlocks {
			err := bk.add(shardNO, group[i])
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if !bk.findAll(blockIDs) {
		t.Fatal("Wrong")
	}

	retrievedHalfBlocks := bk.retrieveAll(halfBlocks)
	for shardNO, blocks := range retrievedHalfBlocks {
		if len(blocks) != 25 {
			t.Fatal("wrong")
		}
		t.Logf("Shard %d is right", shardNO)
	}
}

// Test for SeparateSequence and AggregateSequence
func TestSeparateAndAggregateSequence(t *testing.T) {
	// Generate test data
	blockSequence := []common.Hash{}
	blockAssociated := []uint8{}
	blockGroups := map[uint8][]elements.Block{}

	bk := newBlockCache(shardNumbers_test)

	var length int

	for _, shardNO := range shardNumbers_test {
		blocks := generateBlocks(shardNO)
		blockGroups[shardNO] = blocks
		length = len(blocks)
	}

	for i := 0; i < length; i++ {
		for shardNO, group := range blockGroups {
			err := bk.add(shardNO, group[i])
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	for i := 0; i < length; i++ {
		for _, shardNO := range shardNumbers_test {
			blockSequence = append(blockSequence, blockGroups[shardNO][i].BlockID)
			blockAssociated = append(blockAssociated, shardNO)
		}
	}

	// Separate the sequence
	groupedBlocks, err := SeparateSequence(blockSequence, blockAssociated)
	if err != nil {
		t.Fatalf("SeparateSequence returned an error: %v", err)
	}

	if !bk.findAll(groupedBlocks) {
		t.Fatal("cannot find all blocks")
	}

	// Aggregate the sequence
	aggregatedBlocks, err := AggregateSequence(blockGroups, blockSequence, blockAssociated)
	if err != nil {
		t.Fatalf("AggregateSequence returned an error: %v", err)
	}

	// Verify the result
	for i, blockID := range blockSequence {
		if aggregatedBlocks[i].BlockID != blockID {
			t.Fatalf("BlockID mismatch at index %d: expected %x, got %x", i, blockID, aggregatedBlocks[i].BlockID)
		}
	}
}
