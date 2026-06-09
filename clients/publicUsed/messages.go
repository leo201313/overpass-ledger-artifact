package publicUsed

import (
	"fmt"
	"opl/common"
	"opl/elements"
	"strings"
)

type CDNT_NETWORK_INFO struct {
	NodeName         string
	SelfAddr         string
	Party            string
	ConnectedCDNTs   uint64
	ConnectedWorkers uint64
	WorkerShard      []uint64
}

type CDNT_STATE_INFO struct {
	NodeName   string
	SelfAddr   string
	Party      string
	NowState   uint64
	NowVersion common.Hash
	NowHeight  uint64
}

type WORKER_NETWORK_INFO struct {
	NodeName          string
	SelfAddr          string
	Party             string
	ConnectedCDNTAddr string
	ConnectedWorkers  uint64
	ShardNumber       uint64
}

type WORKER_STATE_INFO struct {
	NodeName    string
	SelfAddr    string
	Party       string
	ShardNumber uint64
	NowState    uint64
	NowVersion  common.Hash
	NowHeight   uint64
}

type TX_GROUP_MSG struct {
	TXS []elements.Transaction
}

type TX_MSG struct {
	TX elements.Transaction
}

type TXIDS_BY_HEIGHT struct {
	GOT     uint64 // 0 is false, 1 is true
	TXTYPES []uint8
	TXIDS   []common.Hash
}

//type TXTYPES_BY_HEIGHT struct {
//	GOT     uint64  // 0 is false, 1 is true
//	TXTYPES []uint8 // 0 is inherited, 1 is re-executed
//}

type DETAIL_INFO struct {
	Content string
}

// String provides a formatted string representation of the CDNT_NETWORK_INFO structure
func (cnt *CDNT_NETWORK_INFO) String() string {
	return fmt.Sprintf(
		"NodeName: %s\nSelfAddr: %s\nParty: %s\nConnectedCDNTs: %d\nConnectedWorkers: %d\nWorkerShard: [%s]",
		cnt.NodeName,
		cnt.SelfAddr,
		cnt.Party,
		cnt.ConnectedCDNTs,
		cnt.ConnectedWorkers,
		strings.Trim(strings.Replace(fmt.Sprint(cnt.WorkerShard), " ", ", ", -1), "[]"),
	)
}

// String provides a formatted string representation of the CDNT_STATE_INFO structure
func (csi *CDNT_STATE_INFO) String() string {
	return fmt.Sprintf(
		"NodeName: %s\nSelfAddr: %s\nParty: %s\nNowState: %d\nNowVersion: %x\nNowHeight: %d",
		csi.NodeName,
		csi.SelfAddr,
		csi.Party,
		csi.NowState,
		csi.NowVersion,
		csi.NowHeight,
	)
}

// String provides a formatted string representation of the WORKER_NETWORK_INFO structure
func (wni *WORKER_NETWORK_INFO) String() string {
	return fmt.Sprintf(
		"NodeName: %s\nSelfAddr: %s\nParty: %s\nConnectedCDNTAddr: %s\nConnectedWorkers: %d\nShardNumber: %d",
		wni.NodeName,
		wni.SelfAddr,
		wni.Party,
		wni.ConnectedCDNTAddr,
		wni.ConnectedWorkers,
		wni.ShardNumber,
	)
}

// String provides a formatted string representation of the WORKER_STATE_INFO structure
func (wsi *WORKER_STATE_INFO) String() string {
	return fmt.Sprintf(
		"NodeName: %s\nSelfAddr: %s\nParty: %s\nShardNumber: %d\nNowState: %d\nNowVersion: %x\nNowHeight: %d",
		wsi.NodeName,
		wsi.SelfAddr,
		wsi.Party,
		wsi.ShardNumber,
		wsi.NowState,
		wsi.NowVersion,
		wsi.NowHeight,
	)
}
