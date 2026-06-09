package entity

import (
	"crypto/sha256"
	rand2 "math/rand/v2"
	"opl/common"
	"opl/database"
	"opl/smartcontract"
	"opl/stateManager"
	"testing"
	"time"
)

//
//func TestWholeWorkFlow(t *testing.T) {
//	cdnts := newTestCoordinators()
//	workers := newTestWorkers()
//	for i := 0; i < len(cdnts); i++ {
//		go cdnts[i].StartDial()
//	}
//	for i := 0; i < len(workers); i++ {
//		go workers[i].StartDial()
//	}
//
//	time.Sleep(3 * time.Second)
//
//	for i := 0; i < len(cdnts); i++ {
//		if err := cdnts[i].BindPeers(); err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	for i := 0; i < len(workers); i++ {
//		if err := workers[i].BindPeers(); err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	for i := 0; i < len(cdnts); i++ {
//		cdnts[i].Run()
//	}
//
//	for i := 0; i < len(workers); i++ {
//		workers[i].Run()
//	}
//
//	//--------------------All coordinators and Workers have begun
//	txAmount := 1000
//	txs_t := generateTxKVStore(shardNO_test, txAmount)
//
//	//--------------------Test1: Block can generate and update, epoch can commit and step in next
//	for i := 0; i < txAmount; i++ {
//		workers[i%len(workers)].GetTransaction(txs_t[i])
//	}
//
//	time.Sleep(5 * time.Second)
//	verCdnt := cdnts[0].stateContainer.nowVersion
//	heightCdnt := cdnts[0].stateContainer.nowHeight
//	verWorker := workers[0].stateContainer.localEpochVersion
//	heightWorker := workers[0].stateContainer.nowHeight
//	if verWorker != verCdnt {
//		t.Fatal("version not match")
//	}
//	t.Logf("The upper version is: %x,  height is %d", verCdnt, heightCdnt)
//	t.Logf("The lower version is: %x,  height is %d", verWorker, heightWorker)
//
//	//if workers[0].stateContainer.lus.Used != true {
//	//	t.Fatal("The tick message is not available")
//	//}
//
//	for i := 0; i < int(heightWorker); i++ {
//		epoch, err := cdnts[0].worldStateManager.GetEpochByHeight(uint64(i + 1))
//		if err != nil {
//			t.Fatal(err)
//		}
//		t.Log(epoch.String())
//	}
//
//	//--------------------Test2: The transactions are properly executed through MVCC and overpass
//	account1 := "Andy"
//	account2 := "Bob"
//	accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(account1)))
//	accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(account2)))
//	checkAddr1 := smartcontract.InvertAsCheckAddr(accountAddr1)
//	checkAddr2 := smartcontract.InvertAsCheckAddr(accountAddr2)
//
//	totalTxAmount := 1000
//	spTxs := generateSmallBankTxsTransfer(account1, account2, totalTxAmount)
//	for i := 0; i < totalTxAmount; i++ {
//		workers[i%len(workers)].GetTransaction(spTxs[i])
//	}
//	time.Sleep(5 * time.Second)
//
//	verCdnt = cdnts[0].stateContainer.nowVersion
//	heightCdnt = cdnts[0].stateContainer.nowHeight
//	verWorker = workers[0].stateContainer.localEpochVersion
//	heightWorker = workers[0].stateContainer.nowHeight
//	if verWorker != verCdnt {
//		t.Fatal("version not match")
//	}
//	t.Logf("The upper version is: %x,  height is %d", verCdnt, heightCdnt)
//	t.Logf("The lower version is: %x,  height is %d", verWorker, heightWorker)
//
//	//if workers[0].stateContainer.lus.Used != true {
//	//	t.Fatal("The tick message is not available")
//	//}
//
//	for i := 0; i < int(heightWorker); i++ {
//		epoch, err := cdnts[0].worldStateManager.GetEpochByHeight(uint64(i + 1))
//		if err != nil {
//			t.Fatal(err)
//		}
//		t.Log(epoch.String())
//	}
//
//	have, val1 := workers[0].worldStateManager.ReadState(checkAddr1)
//	if !have {
//		t.Fatal("Cannot read value")
//	}
//	bal1, err := smartcontract.BytesToInt(val1)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	have, val2 := workers[0].worldStateManager.ReadState(checkAddr2)
//	if !have {
//		t.Fatal("Cannot read value")
//	}
//	bal2, err := smartcontract.BytesToInt(val2)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// The initial balance is 100000
//	t.Logf("Alice now has: %d, Bob now has: %d", bal1, bal2)
//	if bal1+totalTxAmount != bal2-totalTxAmount {
//		t.Fatal("The transactions are not processed properly")
//	}
//
//}
//
//func TestWholeWorkFlow2(t *testing.T) {
//
//	cdnts := newTestCoordinators()
//	workers := newTestWorkers()
//	for i := 0; i < len(cdnts); i++ {
//		go cdnts[i].StartDial()
//	}
//	for i := 0; i < len(workers); i++ {
//		go workers[i].StartDial()
//	}
//
//	time.Sleep(3 * time.Second)
//
//	for i := 0; i < len(cdnts); i++ {
//		if err := cdnts[i].BindPeers(); err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	for i := 0; i < len(workers); i++ {
//		if err := workers[i].BindPeers(); err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	for i := 0; i < len(cdnts); i++ {
//		cdnts[i].Run()
//	}
//
//	for i := 0; i < len(workers); i++ {
//		workers[i].Run()
//	}
//
//	//--------------------All coordinators and Workers have begun
//	txAmount := 1000
//	txs_sb := generateTransactions_SmallBank(txAmount)
//
//	//--------------------Test1: Block can generate and update, epoch can commit and step in next
//	for i := 0; i < txAmount; i++ {
//		workers[i%len(workers)].GetTransaction(txs_sb[i])
//	}
//
//	time.Sleep(5 * time.Second)
//	verCdnt := cdnts[0].stateContainer.nowVersion
//	heightCdnt := cdnts[0].stateContainer.nowHeight
//	verWorker := workers[0].stateContainer.localEpochVersion
//	heightWorker := workers[0].stateContainer.nowHeight
//	if verWorker != verCdnt {
//		t.Fatal("version not match")
//	}
//	t.Logf("The upper version is: %x,  height is %d", verCdnt, heightCdnt)
//	t.Logf("The lower version is: %x,  height is %d", verWorker, heightWorker)
//
//	//if workers[0].stateContainer.lus.Used != true {
//	//	t.Fatal("The tick message is not available")
//	//}
//
//	for i := 0; i < int(heightWorker); i++ {
//		epoch, err := cdnts[0].worldStateManager.GetEpochByHeight(uint64(i + 1))
//		if err != nil {
//			t.Fatal(err)
//		}
//		t.Log(epoch.String())
//
//	}
//
//}

func TestWholeWorkFlow3(t *testing.T) {
	cdnts := newTestCoordinators_()
	workers := newTestWorkers_()
	for i := 0; i < len(cdnts); i++ {
		go cdnts[i].StartDial()
	}
	for i := 0; i < len(workers); i++ {
		go workers[i].StartDial()
	}

	time.Sleep(10 * time.Second)

	for i := 0; i < len(cdnts); i++ {
		if err := cdnts[i].BindPeers(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(workers); i++ {
		if err := workers[i].BindPeers(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(cdnts); i++ {
		cdnts[i].Run()
	}

	for i := 0; i < len(workers); i++ {
		workers[i].Run()
	}

	workerLeaders := make([]*Worker, len(shardNO_test_))
	for i := 0; i < len(shardNO_test_); i++ {
		workerLeaders[i] = workers[i*len(parties_test_)+(rand2.Int()%len(parties_test_))]
	}

	// send many rubbish transactions
	for i := 0; i < len(workerLeaders); i++ {
		go func() {
			for j := 0; j < 5; j++ {
				txs := generateTxKVStore(workerLeaders[i].Shard, 100)
				err := workerLeaders[i].GetTransactionGroup(txs)
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	accAmount := 10000
	totalTxs := 1000

	sendTimes := 3

	for turn := 0; turn < sendTimes; turn++ {
		// send many transfer transactions
		for i := 0; i < len(workerLeaders); i++ {
			go func() {
				txs := generateTransactions_Transfer(0, totalTxs, accAmount)
				err := workerLeaders[i].GetTransactionGroup(txs)
				if err != nil {
					t.Fatal(err)
				}
			}()
		}
		time.Sleep(time.Second)
	}

	// send many rubbish transactions
	for i := 0; i < len(workerLeaders); i++ {
		go func() {
			for j := 0; j < 5; j++ {
				txs := generateTxKVStore(workerLeaders[i].Shard, 100)
				err := workerLeaders[i].GetTransactionGroup(txs)
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	time.Sleep(2 * time.Second)

	mid := accAmount / 2

	for i := 0; i < totalTxs; i++ {
		fromAccount := smartcontract.BackAccountNameByIndex(i)
		toAccount := smartcontract.BackAccountNameByIndex((i + mid) % accAmount)
		fromAddr := common.HashToAddress(sha256.Sum256([]byte(fromAccount)))
		toAddr := common.HashToAddress(sha256.Sum256([]byte(toAccount)))

		have, val1 := workers[0].worldStateManager.ReadState(fromAddr)
		if !have {
			t.Fatalf("Cannot read value of account %s", fromAccount)
		}
		have, val2 := workers[0].worldStateManager.ReadState(toAddr)
		if !have {
			t.Fatalf("Cannot read value of account %s", toAccount)
		}

		bal1, err := smartcontract.BytesToInt(val1)
		if err != nil {
			t.Fatal("Cannot convert bytes")
		}
		bal2, err := smartcontract.BytesToInt(val2)
		if err != nil {
			t.Fatal("Cannot convert bytes")
		}

		if bal1 != smartcontract.BANLANCE-sendTimes*len(shardNO_test_) {
			t.Fatalf("the balance not right, expect: %d, got: %d", smartcontract.BANLANCE-sendTimes*len(shardNO_test_), bal1)
		}

		if bal2 != smartcontract.BANLANCE+sendTimes*len(shardNO_test_) {
			t.Fatalf("the balance not right, expect: %d, got: %d", smartcontract.BANLANCE+sendTimes*len(shardNO_test_), bal2)
		}

	}

}

func TestPeerConnect(t *testing.T) {
	cdnts := newTestCoordinators_()
	workers := newTestWorkers_()
	for i := 0; i < len(cdnts); i++ {
		go cdnts[i].StartDial()
	}
	for i := 0; i < len(workers); i++ {
		go workers[i].StartDial()
	}

	time.Sleep(10 * time.Second)

	for i := 0; i < len(cdnts); i++ {
		if err := cdnts[i].BindPeers(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(workers); i++ {
		if err := workers[i].BindPeers(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(cdnts); i++ {
		cdnts[i].Run()
	}

	for i := 0; i < len(workers); i++ {
		workers[i].Run()
	}

}

var cdntAddrs_test_ = []string{"127.0.0.1:3030", "127.0.0.1:3031", "127.0.0.1:3032", "127.0.0.1:3033"}
var workerAddrs_test_ = [][]string{
	[]string{"127.0.0.1:4030", "127.0.0.1:4031", "127.0.0.1:4032", "127.0.0.1:4033"},
	[]string{"127.0.0.1:4130", "127.0.0.1:4131", "127.0.0.1:4132", "127.0.0.1:4133"},
	[]string{"127.0.0.1:4230", "127.0.0.1:4231", "127.0.0.1:4232", "127.0.0.1:4233"},
	[]string{"127.0.0.1:4330", "127.0.0.1:4331", "127.0.0.1:4332", "127.0.0.1:4333"},
}
var parties_test_ = []string{"Org1", "Org2", "Org3", "Org4"}

var shardNO_test_ = []uint8{1, 2, 3, 4}

func newTestCoordinators_() []*Coordinator {
	coordinators := make([]*Coordinator, len(cdntAddrs_test_))
	for i := 0; i < len(coordinators); i++ {
		memDB := database.NewSimpleMemLDB()
		wsm := stateManager.NewWorldStateManager(memDB)
		sme := smartcontract.NewDemoSME()

		nodeName := "CDNT-" + cdntAddrs_test_[i]
		selfAddr := cdntAddrs_test_[i]
		partyName := parties_test_[i]
		workerAddrs := make([]string, len(shardNO_test_))
		for j := 0; j < len(shardNO_test_); j++ {
			workerAddrs[j] = workerAddrs_test_[j][i]
		}

		cdnt, err := NewCoordinator(nodeName, selfAddr, partyName, cdntAddrs_test_, workerAddrs, shardNO_test_, wsm, sme)
		if err != nil {
			panic(err)
		}
		coordinators[i] = cdnt
	}
	return coordinators
}

func newTestWorkers_() []*Worker {
	workers := make([]*Worker, len(shardNO_test_)*len(parties_test_))
	for i := 0; i < len(workers); i++ {
		memDB := database.NewSimpleMemLDB()
		wsm := stateManager.NewWorldStateManager(memDB)
		sme := smartcontract.NewDemoSME()

		shardNO_index := i / len(parties_test_)
		shardNO := shardNO_test_[shardNO_index]

		nodeName := "WORKER-" + workerAddrs_test_[shardNO_index][i%len(parties_test_)]
		selfAddr := workerAddrs_test_[shardNO_index][i%len(parties_test_)]
		cdntAddr := cdntAddrs_test_[i%len(parties_test_)]
		partyName := parties_test_[i%len(parties_test_)]

		worker, err := NewWorker(nodeName, selfAddr, cdntAddr, partyName, shardNO, workerAddrs_test_[shardNO_index], wsm, sme)
		if err != nil {
			panic(err)
		}

		workers[i] = worker
	}
	return workers
}
