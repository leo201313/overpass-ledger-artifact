package entity

//
//func newCoordinatorForTestManager(nodeName, selfAddr, partyName string, cdntAddrs []string, shardNOs []uint8, db *database.SimpleLDB, managerAddr string) *Coordinator {
//	cdntPeers := make([]string, 0)
//	for _, cp := range cdntAddrs {
//		if cp == selfAddr {
//			continue
//		}
//		cdntPeers = append(cdntPeers, cp)
//	}
//
//	cdntAmount := len(cdntPeers) + 1
//	f := (cdntAmount - 1) / 3
//	th0 := cdntAmount - f
//	th1 := cdntAmount
//
//	host := network.NewNodeHost(nodeName, selfAddr)
//	host.Start()
//
//	wsm := stateManager.NewWorldStateManager(db)
//	sme := smartcontract.NewDemoSME()
//	usm := stateManager.NewSimpleUppStateManager(sme, wsm)
//
//	cdnt := Coordinator{
//		NodeName:          nodeName,
//		SelfAddr:          selfAddr,
//		Party:             partyName,
//		CdntAddrs:         cdntPeers,
//		WorkerAddrs:       nil,
//		WorkerShardNOs:    shardNOs,
//		HostClient:        host,
//		CdntPG:            network.NewPeerGroup(),
//		WorkerPG:          network.NewPeerGroup(),
//		threshold0:        th0,
//		threshold1:        th1,
//		stateContainer:    newCdntState(),
//		blockManager:      nil,
//		uppStateManager:   usm,
//		worldStateManager: wsm,
//		testMode:          true,
//		testContainer:     newTestStateContainer(managerAddr),
//	}
//	return &cdnt
//}
//
//var shardNOs_TestManager = []uint8{0, 1, 2}
//var cdntAddrs_TestManager = []string{"127.0.0.1:3030", "127.0.0.1:3031", "127.0.0.1:3032", "127.0.0.1:3033"}
//var managerAddr_TestManager = "127.0.0.1:3040"
//var parties_TestManager = []string{"Org1", "Org2", "Org3", "Org4"}
//
//func makeCoordinators_TestManager_MemDB() []*Coordinator {
//	cdnts := []*Coordinator{}
//	for i := 0; i < len(cdntAddrs_TestManager); i++ {
//		cdnt := NewTestCoordinator(parties_TestManager[i]+"-CDNT", cdntAddrs_TestManager[i], parties_TestManager[i], cdntAddrs_TestManager, shardNOs_TestManager, database.NewSimpleMemLDB(), managerAddr_TestManager)
//		cdnts = append(cdnts, cdnt)
//	}
//	return cdnts
//}
//
//func makeManager_TestManager_MemDB() *TestManager {
//	return NewTestManager(managerAddr_TestManager, cdntAddrs_TestManager, shardNOs_TestManager, database.NewSimpleMemLDB())
//}
//
//func TestTestManager_MemDB(t *testing.T) {
//
//	printTestStateMessage := func(msg TestStateMsg) {
//		text := fmt.Sprintf("The state code is %d \n", msg.State)
//		text += fmt.Sprintf("Time used %d mms for %d transactions in %d blocks\n", msg.TimeUsed, msg.TxNum, msg.BlockNum)
//		deltaTime := float64(msg.TimeUsed) / 1000000
//		tps := float64(msg.TxNum) / deltaTime
//		text += fmt.Sprintf("The tps is: %.4f", tps)
//		t.Log(text)
//
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
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
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
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
//	if err != nil {
//		t.Fatal(err)
//	}
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
//
//var db_dir = "./testDBs"
//
//func generateDB(rank int) *database.SimpleLDB {
//	dbName := "db_" + strconv.Itoa(rank)
//	dbPath := db_dir + "/" + dbName
//	err := os.RemoveAll(dbPath)
//	if err != nil {
//		panic(err)
//	}
//	db, err := leveldb.OpenFile(dbPath, nil)
//	if err != nil {
//		panic(err)
//	}
//	return database.NewSimpleLDB(dbName, db)
//}
//
//func openDB(rank int) *database.SimpleLDB {
//	dbName := "db_" + strconv.Itoa(rank)
//	dbPath := db_dir + "/" + dbName
//	db, err := leveldb.OpenFile(dbPath, nil)
//	if err != nil {
//		panic(err)
//	}
//	return database.NewSimpleLDB(dbName, db)
//}
//
//var huge_amount = 10000000
//
//func putHugeData(db *database.SimpleLDB) {
//	for i := 0; i < huge_amount; i++ {
//		key := common.BytesToAddress([]byte(strconv.Itoa(i)))
//		value := []byte("123")
//		err := db.Put(key[:], value)
//		if err != nil {
//			panic(err)
//		}
//	}
//}
//
//func makeCoordinators_TestManager_STDB() []*Coordinator {
//	cdnts := []*Coordinator{}
//	for i := 0; i < len(cdntAddrs_TestManager); i++ {
//		db := openDB(i)
//		cdnt := NewTestCoordinator(parties_TestManager[i]+"-CDNT", cdntAddrs_TestManager[i], parties_TestManager[i], cdntAddrs_TestManager, shardNOs_TestManager, db, managerAddr_TestManager)
//		cdnts = append(cdnts, cdnt)
//	}
//	return cdnts
//}
//
//func makeManager_TestManager_STDB() *TestManager {
//	db := openDB(len(cdntAddrs_TestManager))
//
//	return NewTestManager(managerAddr_TestManager, cdntAddrs_TestManager, shardNOs_TestManager, db)
//}
//
//func generateTxKVStore_STDB(shardNO uint8, amount int) []elements.Transaction {
//	res := make([]elements.Transaction, amount)
//
//	base := huge_amount / len(shardNOs_TestManager)
//	basebase := base / 2
//
//	for i := 0; i < amount; i++ {
//		r1 := rand.Int() % 3
//		r2 := rand.Int()%basebase + int(shardNO)*base
//		accountAddr := common.BytesToAddress([]byte(strconv.Itoa(r2)))
//		switch r1 {
//		case 0: // read
//
//			arg1 := elements.Argument{
//				Type:    elements.ADDR_ARG,
//				Address: accountAddr,
//				Value:   nil,
//			}
//
//			tx := elements.Transaction{
//				TxID:          common.Hash{},
//				Sender:        common.Address{},
//				Version:       common.GenerateRandomHash(),
//				Nonce:         0,
//				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
//				Function:      smartcontract.KVSTORE_FUNC_READ,
//				Arguments:     []elements.Argument{arg1},
//				Signature:     nil,
//				StateReadSet:  nil,
//				StateWriteSet: nil,
//				Results:       nil,
//			}
//
//			tx.SetTxID()
//			res[i] = tx
//
//		case 1: // write
//
//			arg1 := elements.Argument{
//				Type:    elements.ADDR_ARG,
//				Address: accountAddr,
//				Value:   nil,
//			}
//
//			r3 := rand.Int() % 100000
//			arg2 := elements.Argument{
//				Type:    elements.VALUE_ARG,
//				Address: common.Address{},
//				Value:   smartcontract.IntToBytes(r3),
//			}
//
//			tx := elements.Transaction{
//				TxID:          common.Hash{},
//				Sender:        common.Address{},
//				Version:       common.GenerateRandomHash(),
//				Nonce:         0,
//				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
//				Function:      smartcontract.KVSTORE_FUNC_WRITE,
//				Arguments:     []elements.Argument{arg1, arg2},
//				Signature:     nil,
//				StateReadSet:  nil,
//				StateWriteSet: nil,
//				Results:       nil,
//			}
//
//			tx.SetTxID()
//			res[i] = tx
//
//		case 2: //delete
//
//			arg1 := elements.Argument{
//				Type:    elements.ADDR_ARG,
//				Address: accountAddr,
//				Value:   nil,
//			}
//
//			tx := elements.Transaction{
//				TxID:          common.Hash{},
//				Sender:        common.Address{},
//				Version:       common.GenerateRandomHash(),
//				Nonce:         0,
//				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
//				Function:      smartcontract.KVSTORE_FUNC_DELETE,
//				Arguments:     []elements.Argument{arg1},
//				Signature:     nil,
//				StateReadSet:  nil,
//				StateWriteSet: nil,
//				Results:       nil,
//			}
//			tx.SetTxID()
//			res[i] = tx
//		}
//
//	}
//	return res
//}
//
//func generateTxKVStore_STDB_AllRead(shardNO uint8, amount int) []elements.Transaction {
//
//	res := make([]elements.Transaction, amount)
//
//	base := huge_amount / len(shardNOs_TestManager)
//	basebase := base / 2
//
//	for i := 0; i < amount; i++ {
//		r2 := rand.Int()%basebase + int(shardNO)*base
//		accountAddr := common.BytesToAddress([]byte(strconv.Itoa(r2)))
//
//		arg1 := elements.Argument{
//			Type:    elements.ADDR_ARG,
//			Address: accountAddr,
//			Value:   nil,
//		}
//
//		tx := elements.Transaction{
//			TxID:          common.Hash{},
//			Sender:        common.Address{},
//			Version:       common.GenerateRandomHash(),
//			Nonce:         0,
//			Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
//			Function:      smartcontract.KVSTORE_FUNC_READ,
//			Arguments:     []elements.Argument{arg1},
//			Signature:     nil,
//			StateReadSet:  nil,
//			StateWriteSet: nil,
//			Results:       nil,
//		}
//
//		tx.SetTxID()
//		res[i] = tx
//	}
//	return res
//}
//
//func TestTestManager_MakeDBs(t *testing.T) {
//	for i := 0; i < len(cdntAddrs_TestManager); i++ {
//		db := generateDB(i)
//		putHugeData(db)
//		db.Close()
//	}
//	db := generateDB(len(cdntAddrs_TestManager))
//	putHugeData(db)
//	db.Close()
//}
//
//func TestTestManager_STDB(t *testing.T) {
//
//	printTestStateMessage := func(msg TestStateMsg) {
//		text := fmt.Sprintf("The state code is %d \n", msg.State)
//		text += fmt.Sprintf("Time used %d mms for %d transactions in %d blocks\n", msg.TimeUsed, msg.TxNum, msg.BlockNum)
//		deltaTime := float64(msg.TimeUsed) / 1000000
//		tps := float64(msg.TxNum) / deltaTime
//		text += fmt.Sprintf("The tps is: %.4f", tps)
//		t.Log(text)
//
//	}
//
//	cdnts := makeCoordinators_TestManager_STDB()
//	tm := makeManager_TestManager_STDB()
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
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore_STDB(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
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
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore_STDB(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
//	if err != nil {
//		t.Fatal(err)
//	}
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
//
//}
//
//func TestTestManager_STDB_AllRead(t *testing.T) {
//	printTestStateMessage := func(msg TestStateMsg) {
//		text := fmt.Sprintf("The state code is %d \n", msg.State)
//		text += fmt.Sprintf("Time used %d mms for %d transactions in %d blocks\n", msg.TimeUsed, msg.TxNum, msg.BlockNum)
//		deltaTime := float64(msg.TimeUsed) / 1000000
//		tps := float64(msg.TxNum) / deltaTime
//		text += fmt.Sprintf("The tps is: %.4f", tps)
//		t.Log(text)
//
//	}
//
//	cdnts := makeCoordinators_TestManager_STDB()
//	tm := makeManager_TestManager_STDB()
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
//	// test transactions on KV-STORE contract Read function with all WO transaction
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore_STDB_AllRead(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	time.Sleep(5 * time.Second)
//	tm.QueryState(0)
//
//	time.Sleep(2 * time.Second)
//
//	if len(tm.savedStateMessage) != 1 {
//		t.Fatal()
//	}
//
//	printTestStateMessage(tm.savedStateMessage[0])
//
//	// test transactions on KV-STORE contract Read function with all WP transaction
//	for round := 0; round < 1; round++ {
//		for i := 0; i < len(shardNOs_TestManager); i++ {
//			txs := generateTxKVStore_STDB_AllRead(shardNOs_TestManager[i], 10000)
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
//	err = tm.TriggerUppConsensus(0)
//	if err != nil {
//		t.Fatal(err)
//	}
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
