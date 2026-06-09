package stateManager

import (
	"crypto/sha256"
	"fmt"
	"opl/common"
	"opl/database"
	"opl/elements"
	"opl/smartcontract"
	"testing"
)

var (
	testShardNo1 uint8 = 0
	testShardNo2 uint8 = 1
)

func TestBothSimpleStateManager(t *testing.T) {

	printFunc := func(wsmName string, accountName string, isSaving bool, wsm *WorldStateManager, want int) {
		text := "Query from " + wsmName + ":\n"
		if isSaving {
			got := queryAccountFromWSM(accountName, isSaving, wsm)
			if got != want {
				t.Fatalf("The saving account of %s should be %d but got %d", accountName, want, got)
			}
			text += fmt.Sprintf("The saving account of %s is %d", accountName, got)
		} else {
			got := queryAccountFromWSM(accountName, isSaving, wsm)
			if got != want {
				t.Fatalf("The check account of %s should be %d but got %d", accountName, want, got)
			}
			text += fmt.Sprintf("The check account of %s is %d", accountName, got)
		}
		t.Log(text)
	}

	wsmUpp := NewWorldStateManager(database.NewSimpleMemLDB())
	wsmLow1 := NewWorldStateManager(database.NewSimpleMemLDB())
	wsmLow2 := NewWorldStateManager(database.NewSimpleMemLDB())

	smeUpp := smartcontract.NewDemoSME()
	smeLow1 := smartcontract.NewDemoSME()
	smeLow2 := smartcontract.NewDemoSME()

	usm := NewSimpleUppStateManager(smeUpp, wsmUpp)
	lsm1 := NewSimpleLowStateManager(wsmLow1, smeLow1, testShardNo1)
	lsm2 := NewSimpleLowStateManager(wsmLow2, smeLow2, testShardNo2)

	// Alice Bob Cyan David are associated with shard 0
	// Krad Leo Mike Niko are associated with shard 1

	testBlock1 := elements.Block{
		BlockID: common.Hash{},
		ShardNO: testShardNo1,
		Version: common.INITIAL_VERSION,
		Nonce:   0,
		Transactions: []elements.Transaction{
			smartcontract.GenerateTestUpdateBalanceTx("Alice", 100),
			smartcontract.GenerateTestUpdateBalanceTx("Bob", 200),
			smartcontract.GenerateTestSendPaymentTx("Alice", "Bob", 20),
			smartcontract.GenerateTestAlmagate("Cyan", "Mike"),
		},
	}
	testBlock1.SetBlockID()

	testBlock2 := elements.Block{
		BlockID: common.Hash{},
		ShardNO: testShardNo1,
		Version: common.INITIAL_VERSION,
		Nonce:   1,
		Transactions: []elements.Transaction{
			smartcontract.GenerateTestUpdateBalanceTx("Cyan", 300),
			smartcontract.GenerateTestUpdateBalanceTx("David", 500),
			smartcontract.GenerateTestSendPaymentTx("Alice", "Mike", 50),
		},
	}
	testBlock2.SetBlockID()

	testBlock3 := elements.Block{
		BlockID: common.Hash{},
		ShardNO: testShardNo2,
		Version: common.INITIAL_VERSION,
		Nonce:   0,
		Transactions: []elements.Transaction{
			smartcontract.GenerateTestUpdateBalanceTx("Krad", 450),
			smartcontract.GenerateTestUpdateBalanceTx("Leo", 700),
			smartcontract.GenerateTestSendPaymentTx("Krad", "Alice", 50),
		},
	}
	testBlock3.SetBlockID()

	// start test
	wrappedBlock1 := lsm1.ProcessBlock(testBlock1.BlockID, testBlock1.Transactions)
	wrappedBlock2 := lsm1.ProcessBlock(testBlock2.BlockID, testBlock2.Transactions)
	wrappedBlock3 := lsm2.ProcessBlock(testBlock3.BlockID, testBlock3.Transactions)

	blocks := []elements.Block{wrappedBlock1, wrappedBlock3, wrappedBlock2}

	testEpoch := elements.Epoch{
		EpochID:                 common.Hash{},
		PreviousEpoch:           common.INITIAL_VERSION,
		Height:                  1,
		BlockSequence:           []common.Hash{wrappedBlock1.BlockID, wrappedBlock3.BlockID, wrappedBlock2.BlockID},
		BlockAssociatedShardNOs: []uint8{wrappedBlock1.ShardNO, wrappedBlock3.ShardNO, wrappedBlock2.ShardNO},
		Receipts:                nil,
		StateCommitSet:          nil,
	}

	testEpoch.SetEpochID()

	usm.StepNextEpoch(testEpoch.EpochID)
	receipts := usm.ProcessEpoch(blocks)
	stateCommitSet := usm.ExportStateCommitSet()
	usm.CommitNowVCT()

	testEpoch.Receipts = receipts
	testEpoch.StateCommitSet = stateCommitSet

	wsmUpp.CommitStateSet(stateCommitSet)
	wsmLow1.CommitStateSet(stateCommitSet)
	wsmLow2.CommitStateSet(stateCommitSet)

	lsm1.StepNextEpoch(testEpoch)
	lsm2.StepNextEpoch(testEpoch)

	printFunc("LSM1", "Alice", false, wsmLow1, 100080)
	printFunc("LSM1", "Bob", false, wsmLow1, 100220)
	printFunc("LSM1", "Cyan", false, wsmLow1, 300)
	printFunc("LSM1", "David", false, wsmLow1, 100500)
	printFunc("LSM2", "Krad", false, wsmLow2, 100400)
	printFunc("LSM2", "Leo", false, wsmLow2, 100700)
	printFunc("LSM2", "Mike", true, wsmLow2, 200000)
	printFunc("LSM2", "Mike", false, wsmLow2, 100050)

	testBlock4 := elements.Block{
		BlockID: common.Hash{},
		ShardNO: testShardNo2,
		Version: testEpoch.EpochID,
		Nonce:   0,
		Transactions: []elements.Transaction{
			smartcontract.GenerateTestUpdateBalanceTx("Mike", 450),
			smartcontract.GenerateTestUpdateBalanceTx("Leo", 700),
			smartcontract.GenerateTestSendPaymentTx("Leo", "Niko", 50),
		},
	}

	testBlock4.SetBlockID()

	wrappedBlock4 := lsm2.ProcessBlock(testBlock4.BlockID, testBlock4.Transactions)

	testEpoch_1 := elements.Epoch{
		EpochID:                 common.Hash{},
		PreviousEpoch:           testEpoch.EpochID,
		Height:                  2,
		BlockSequence:           []common.Hash{wrappedBlock4.BlockID},
		BlockAssociatedShardNOs: []uint8{wrappedBlock4.ShardNO},
		Receipts:                nil,
		StateCommitSet:          nil,
	}

	usm.StepNextEpoch(testEpoch_1.EpochID)
	receipts = usm.ProcessEpoch([]elements.Block{wrappedBlock4})
	stateCommitSet = usm.ExportStateCommitSet()
	usm.CommitNowVCT()

	testEpoch_1.Receipts = receipts
	testEpoch_1.StateCommitSet = stateCommitSet

	wsmUpp.CommitStateSet(stateCommitSet)
	wsmLow1.CommitStateSet(stateCommitSet)
	wsmLow2.CommitStateSet(stateCommitSet)

	lsm1.StepNextEpoch(testEpoch_1)
	lsm2.StepNextEpoch(testEpoch_1)

	// all world state manager should be the same
	printFunc("LSM2", "Leo", false, wsmLow2, 101350)
	printFunc("LSM2", "Mike", false, wsmLow2, 100500)
	printFunc("LSM2", "Niko", false, wsmLow2, 100050)

	printFunc("LSM1", "Leo", false, wsmLow1, 101350)
	printFunc("LSM1", "Mike", false, wsmLow1, 100500)
	printFunc("LSM1", "Niko", false, wsmLow1, 100050)

	printFunc("USM", "Leo", false, wsmUpp, 101350)
	printFunc("USM", "Mike", false, wsmUpp, 100500)
	printFunc("USM", "Niko", false, wsmUpp, 100050)
}

func queryAccountFromWSM(account string, isSaving bool, wsm *WorldStateManager) int {
	accountAddr := common.HashToAddress(sha256.Sum256([]byte(account)))
	checkAddr := smartcontract.InvertAsCheckAddr(accountAddr)
	if isSaving {
		_, amountBytes := wsm.ReadState(accountAddr)
		amount, _ := smartcontract.BytesToInt(amountBytes)
		return amount
	} else {
		_, amountBytes := wsm.ReadState(checkAddr)
		amount, _ := smartcontract.BytesToInt(amountBytes)
		return amount
	}
}
