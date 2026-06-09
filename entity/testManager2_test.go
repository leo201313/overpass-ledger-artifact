package entity

//
//import (
//	"fmt"
//	"testing"
//	"time"
//)
//
//func TestTestManager_MemDB_UseLoopTrigger(t *testing.T) {
//
//	printTestStateMessage := func(msg TestStateMsg) {
//		text := fmt.Sprintf("The state code is %d \n", msg.State)
//		text += fmt.Sprintf("Time used %d mms for %d transactions in %d blocks\n", msg.TimeUsed, msg.TxNum, msg.BlockNum)
//		deltaTime := float64(msg.TimeUsed) / 1000000
//		tps := float64(msg.TxNum) / deltaTime
//		text += fmt.Sprintf("The tps is: %.4f", tps)
//		t.Log(text)
//	}
//
//	cdnts := makeCoordinators_TestManager_MemDB()
//	tm := makeManager_TestManager_MemDB()
//
//	for i := 0; i < len(cdnts); i++ {
//		go cdnts[i].StartDial()
//	}
//
//	tm.StartDial()
//
//	time.Sleep(2 * time.Second)
//
//	for i := 0; i < len(cdnts); i++ {
//		if err := cdnts[i].BindPeers(); err != nil {
//			t.Fatal(err)
//		}
//	}
//	err := tm.BindPeers()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for i := 0; i < len(cdnts); i++ {
//		cdnts[i].Run()
//	}
//
//	// -------- all nodes have begun
//
//	// test transactions on KV-STORE contract with all WO transaction
//	for round := 0; round < 10; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore(shardNOs_TestManager[i], 1024)
//			err := tm.GenerateBlock(shardNOs_TestManager[i], txs)
//			if err != nil {
//				t.Fatal(err)
//			}
//		}
//	}
//
//	tm.UploadAllBlocks()
//
//	time.Sleep(1 * time.Second)
//
//	err = tm.LoopTriggerOn(0)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	time.Sleep(2 * time.Second)
//	tm.QueryState(0)
//
//	time.Sleep(2 * time.Second)
//
//	if len(tm.savedStateMessage) == 0 {
//		t.Fatal()
//	}
//
//	printTestStateMessage(tm.savedStateMessage[0])
//
//	// test transactions on KV-STORE contract with all WP transaction
//	for round := 0; round < 10; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore(shardNOs_TestManager[i], 1024)
//			err := tm.GenerateBlock(shardNOs_TestManager[i], txs)
//			if err != nil {
//				t.Fatal(err)
//			}
//		}
//	}
//
//	tm.UploadAllBlocks()
//
//	time.Sleep(5 * time.Second)
//	tm.QueryState(0)
//
//	time.Sleep(2 * time.Second)
//
//	if len(tm.savedStateMessage) != 2 {
//		t.Fatal()
//	}
//
//	printTestStateMessage(tm.savedStateMessage[1])
//}
