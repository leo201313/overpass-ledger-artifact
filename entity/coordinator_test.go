package entity

import (
	"opl/common"
	"opl/database"
	"opl/elements"
	"opl/network"
	"opl/smartcontract"
	"opl/stateManager"
	"opl/utils"
	"testing"
	"time"
)

type coordinator_cfg struct {
	nodeName  string
	selfAddr  string
	partyName string
	cdntAddrs []string
	shardNOs  []uint8
}

func makeCoordinator(cfg coordinator_cfg) *Coordinator {
	return newTestCoordinatorForUpperLayer(cfg.nodeName, cfg.selfAddr, cfg.partyName, cfg.cdntAddrs, cfg.shardNOs)
}

func newTestCoordinatorForUpperLayer(nodeName, selfAddr, partyName string, cdntAddrs []string, shardNOs []uint8) *Coordinator {
	cdntPeers := make([]string, 0)
	for _, cp := range cdntAddrs {
		if cp == selfAddr {
			continue
		}
		cdntPeers = append(cdntPeers, cp)
	}

	cdntAmount := len(cdntPeers) + 1
	f := (cdntAmount - 1) / 3
	th0 := cdntAmount - f
	th1 := cdntAmount

	selfRank := utils.ComputeIndexFromIPAddr(selfAddr, cdntPeers)

	host := network.NewNodeHost(nodeName, selfAddr)
	host.Start()

	memDB := database.NewSimpleMemLDB()
	wsm := stateManager.NewWorldStateManager(memDB)
	sme := smartcontract.NewDemoSME()
	usm := stateManager.NewSimpleUppStateManager(sme, wsm)

	cdnt := Coordinator{
		NodeName:          nodeName,
		SelfAddr:          selfAddr,
		Party:             partyName,
		CdntAddrs:         cdntPeers,
		WorkerAddrs:       nil,
		WorkerShardNOs:    shardNOs,
		HostClient:        host,
		CdntPG:            network.NewPeerGroup(),
		WorkerPG:          network.NewPeerGroup(),
		threshold0:        th0,
		threshold1:        th1,
		stateContainer:    newCdntState(selfRank, cdntAmount),
		blockManager:      nil,
		uppStateManager:   usm,
		worldStateManager: wsm,
	}
	return &cdnt
}

var shardNOs_test = []uint8{1, 2, 3}

func makeCoordinators() []*Coordinator {
	cdnts := []*Coordinator{}
	for i := 0; i < len(cdntAddrs_test); i++ {
		tempCfg := coordinator_cfg{
			nodeName:  parties_test[i] + "-CDNT",
			selfAddr:  cdntAddrs_test[i],
			partyName: parties_test[i],
			cdntAddrs: cdntAddrs_test,
			shardNOs:  shardNOs_test,
		}
		cdnt := makeCoordinator(tempCfg)
		cdnts = append(cdnts, cdnt)
	}
	return cdnts
}

type blockGenrator struct {
	shardNOs   []uint8
	nonces     []uint64
	nowVersion common.Hash
}

func newBlockGenrator(shardNOs []uint8, baseVersion common.Hash) *blockGenrator {
	nonces := make([]uint64, len(shardNOs))
	return &blockGenrator{
		shardNOs:   shardNOs,
		nonces:     nonces,
		nowVersion: baseVersion,
	}
}

func (bg *blockGenrator) generateRandomBlocks(amount int) []elements.Block {
	blocks := make([]elements.Block, amount*len(bg.shardNOs))
	for i := 0; i < amount; i++ {
		for j := 0; j < len(bg.shardNOs); j++ {

			tempBlock := elements.Block{
				BlockID:      common.Hash{},
				ShardNO:      bg.shardNOs[j],
				Version:      bg.nowVersion,
				Nonce:        bg.nonces[j],
				Transactions: nil,
			}
			tempBlock.SetBlockID()

			blocks[i*len(bg.shardNOs)+j] = tempBlock
			bg.nonces[j] += 1
		}
	}
	return blocks
}

func (bg *blockGenrator) updateVersion(nextVersion common.Hash) {
	bg.nowVersion = nextVersion
	bg.nonces = make([]uint64, len(bg.shardNOs))
}

func TestCoordinator(t *testing.T) {
	cdnts := makeCoordinators()

	for i := 0; i < len(cdnts); i++ {
		go cdnts[i].StartDial()
	}

	time.Sleep(1 * time.Second)

	for i := 0; i < len(cdnts); i++ {
		if err := cdnts[i].BindPeers(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(cdnts); i++ {
		cdnts[i].Run()
	}
	//--------------------All coordinators have begun

	blkG := newBlockGenrator(shardNOs_test, common.INITIAL_VERSION)
	blkLst0 := blkG.generateRandomBlocks(1)

	//--------------------Test1: Standard process
	for _, cdnt := range cdnts {
		for _, blk := range blkLst0 {
			cdnt.blockManager.addBlock(blk)
		}
	}

	time.Sleep(3 * time.Second)
	if cdnts[0].stateContainer.nowHeight == 0 {
		t.Fatal("Epoch height not right")
	}

	t.Logf("The epoch height now is: %d", cdnts[0].stateContainer.nowHeight)

	epoch, err := cdnts[0].worldStateManager.GetEpochByHeight(cdnts[0].stateContainer.nowHeight)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(epoch.String())

	blkG.updateVersion(cdnts[0].stateContainer.nowVersion)

	//--------------------Test2: Whether can step in next epoch
	blkG.updateVersion(cdnts[0].stateContainer.nowVersion)
	blkLst1 := blkG.generateRandomBlocks(1)

	for _, cdnt := range cdnts {
		for _, blk := range blkLst1 {
			cdnt.blockManager.addBlock(blk)
		}
	}

	time.Sleep(3 * time.Second)
	if cdnts[0].stateContainer.nowHeight == 0 {
		t.Fatal("Epoch height not right")
	}

	t.Logf("The epoch height now is: %d", cdnts[0].stateContainer.nowHeight)

	epoch, err = cdnts[0].worldStateManager.GetEpochByHeight(cdnts[0].stateContainer.nowHeight)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(epoch.String())

	blkG.updateVersion(cdnts[0].stateContainer.nowVersion)

}

func TestEqualOrderBlocks(t *testing.T) {
	var blockMap = map[uint8][]common.Hash{
		0: []common.Hash{common.BytesToHash([]byte("01")), common.BytesToHash([]byte("02")), common.BytesToHash([]byte("03"))},
		1: []common.Hash{common.BytesToHash([]byte("11")), common.BytesToHash([]byte("12")), common.BytesToHash([]byte("13")), common.BytesToHash([]byte("14"))},
		2: []common.Hash{common.BytesToHash([]byte("21")), common.BytesToHash([]byte("22")), common.BytesToHash([]byte("23"))},
		3: []common.Hash{common.BytesToHash([]byte("31")), common.BytesToHash([]byte("32"))},
		4: []common.Hash{common.BytesToHash([]byte("41")), common.BytesToHash([]byte("42")), common.BytesToHash([]byte("43"))},
	}
	orderedBlocks, shardNos := EqualOrderBlocks(blockMap)
	if len(orderedBlocks) != 15 {
		t.Fatal("here")
	}
	if len(shardNos) != 15 {
		t.Fatal("here")
	}

	t.Log(shardNos)

}
