package entity

import (
	"fmt"
	"opl/database"
	"opl/elements"
	"opl/smartcontract"
	"opl/stateManager"
	"testing"
	"time"
)

func newTestWorkers() []*Worker {
	workers := make([]*Worker, len(workerAddrs_test))
	for i := 0; i < len(workers); i++ {
		memDB := database.NewSimpleMemLDB()
		wsm := stateManager.NewWorldStateManager(memDB)
		sme := smartcontract.NewDemoSME()

		nodeName := "WORKER-" + workerAddrs_test[i]
		selfAddr := workerAddrs_test[i]
		cdntAddr := cdntAddrs_test[i]
		partyName := parties_test[i]

		worker, err := NewWorker(nodeName, selfAddr, cdntAddr, partyName, shardNO_test, workerAddrs_test, wsm, sme)
		if err != nil {
			panic(err)
		}

		workers[i] = worker
	}
	return workers
}

func newTestCoordinators() []*Coordinator {
	coordinators := make([]*Coordinator, len(cdntAddrs_test))
	for i := 0; i < len(coordinators); i++ {
		memDB := database.NewSimpleMemLDB()
		wsm := stateManager.NewWorldStateManager(memDB)
		sme := smartcontract.NewDemoSME()

		nodeName := "CDNT-" + cdntAddrs_test[i]
		selfAddr := cdntAddrs_test[i]
		partyName := parties_test[i]
		workerAddrs := []string{workerAddrs_test[i]}

		cdnt, err := NewCoordinator(nodeName, selfAddr, partyName, cdntAddrs_test, workerAddrs, []uint8{shardNO_test}, wsm, sme)
		if err != nil {
			panic(err)
		}
		coordinators[i] = cdnt
	}
	return coordinators
}

func TestWorker(t *testing.T) {
	cdnts := newTestCoordinators()
	workers := newTestWorkers()
	for i := 0; i < len(cdnts); i++ {
		go cdnts[i].StartDial()
	}
	for i := 0; i < len(workers); i++ {
		go workers[i].StartDial()
	}

	time.Sleep(3 * time.Second)

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

	//--------------------All coordinators and Workers have begun
	txAmount := 1
	txs_t := generateTxKVStore(shardNO_test, txAmount)

	//--------------------Test1: Block can generate and update
	for i := 0; i < txAmount; i++ {
		workers[i%len(workers)].GetTransaction(txs_t[i])
	}

	time.Sleep(3 * time.Second)
	blocks := cdnts[0].blockManager.readAll()
	txCount := printBlocks(blocks)
	if txCount != len(txs_t) {
		t.Fatal("Transactions not match")
	}

	//--------------------Test2: Epoch can be distributed and update
	cdnts[0].TryLaunchEpoch()
	time.Sleep(3 * time.Second)
	if workers[0].stateContainer.nowHeight == 0 {
		t.Fatal("fail in epoch update")
	}
	t.Logf("Now the local epoch version is %x and height is %d", workers[0].stateContainer.localEpochVersion, workers[0].stateContainer.nowHeight)

	//--------------------Test3: After Epoch is updated, generate new block
	txs_t = generateTxKVStore(shardNO_test, txAmount)

	for i := 0; i < txAmount; i++ {
		workers[i%len(workers)].GetTransaction(txs_t[i])
	}

	time.Sleep(3 * time.Second)
	blocks = cdnts[0].blockManager.readAll()
	txCount = printBlocks(blocks)
	if txCount != (len(txs_t)) {
		t.Fatal("Transactions not match")
	}

	//--------------------Test4: Compute the total delay
	cdnts[0].TryLaunchEpoch()
	txs := generateTxKVStore(shardNO_test, 1)
	tempTx := txs[0]

	nowNonce := workers[0].stateContainer.nextNonce
	nowHeight := workers[0].stateContainer.nowHeight

	endFlag := make(chan struct{})

	var time1, time2, time3 time.Time

	time1 = time.Now()

	workers[0].GetTransaction(tempTx)
	go func() {
		tiker := time.NewTicker(100 * time.Millisecond)
		for {
			<-tiker.C
			if workers[0].stateContainer.nextNonce != nowNonce {
				time2 = time.Now()
				cdnts[0].TryLaunchEpoch()
				break
			}
		}
	}()
	go func() {
		tiker := time.NewTicker(100 * time.Millisecond)
		for {
			<-tiker.C
			if workers[0].stateContainer.nowHeight != nowHeight {
				time3 = time.Now()
				endFlag <- struct{}{}
				break
			}
		}
	}()

	<-endFlag
	delta_lower := time2.Sub(time1)
	delta_upper := time3.Sub(time2)
	delta_tot := time3.Sub(time1)
	t.Logf("Delta Lower: %v, Delta Upper: %v, Total Delta: %v", delta_lower, delta_upper, delta_tot)

}

// print the blocks and back the number of total transactions
func printBlocks(blocks map[uint8][]elements.Block) int {
	count := 0
	for shardNO, lst := range blocks {
		fmt.Printf("In shard %d, %d blocks are got:\n", shardNO, len(lst))
		for i, blk := range lst {
			fmt.Printf("Block %d:\n", i)
			fmt.Printf("         BlockID: %x\n", blk.BlockID)
			fmt.Printf("         Version: %x\n", blk.Version)
			fmt.Printf("         Nonce  : %d\n", blk.Nonce)
			fmt.Printf("         ShardNO: %d\n", blk.ShardNO)
			fmt.Printf("         TxNum  : %d\n", len(blk.Transactions))
			count += len(blk.Transactions)
		}
	}
	return count
}
