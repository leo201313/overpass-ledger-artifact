package coes

import (
	"crypto/elliptic"
	"time"
)

// for cryptos and key
// -----------------------------------------------------
var (
	DEFAULT_CURVE = elliptic.P256()
)

const (
	PUBKEY_LENGTH   = 64 // bytes
	PUBKEY_SEPARATE = PUBKEY_LENGTH / 2
	SIG_LENGTH      = 64 // bytes
	SIG_SEPARATE    = SIG_LENGTH / 2
)

// -----------------------------------------------------

// for networks
// -----------------------------------------------------
var (
	ENCRYPT_TRANSPORT_AMONG_ORGANIZIATIONS = true
	ENCRYPT_TRANSPORT_IN_ORGANIZIATIONS    = false
)

const (
	// Maximum time allowed for reading a complete message.
	// This is effectively the amount of time a connection can be idle.
	FrameReadTimeout = 30 * time.Second

	// Maximum amount of time allowed for writing a complete message.
	FrameWriteTimeout = 20 * time.Second

	DefaultDialTimeout = 2 * time.Second

	PingInterval = 15 * time.Second

	ReTryDialWait = 1 * time.Second
)

// for srlpx
const (
	// total timeout for encryption handshake and protocol
	// handshake in both directions.
	HandshakeTimeout = 2 * time.Second

	// This is the timeout for sending the disconnect reason.
	// This is shorter than the usual timeout because we don't want
	// to wait if the connection is known to be bad anyway.
	DiscWriteTimeout = 1 * time.Second
)

// -----------------------------------------------------

// for coordinators
// -----------------------------------------------------
const (
	StateCheckInterval = 10 * time.Millisecond

	// only used for test mode
	TriggerScanInterval = 100 * time.Millisecond
)

// -----------------------------------------------------

// for worker
// -----------------------------------------------------
const (
	WorkerStateCheckInterval = 10 * time.Millisecond

	LeastTransaction = 1     // the least transaction amount for launching a new ronud of consensus
	MaxTransaction   = 10000 // the max transactions allowed in a block
)

// -----------------------------------------------------

// for blockchain stored in db
// -----------------------------------------------------
const (
	EpochPrefix = "EPOCH"
	EpochHeight = "EPOCH_HEIGHT"
)

var EpochHeightBytes = []byte(EpochHeight)

// -----------------------------------------------------

// for test manager
// -----------------------------------------------------
const (
	BlockUploadTriggerInterval = 200 * time.Millisecond
	GenerateBlockInterval      = 500 * time.Millisecond
)

// -----------------------------------------------------

// for block cache
// -----------------------------------------------------
// StrictMaxWaitDelay defines the maximum time the coordinator will wait for
// a new consensus launch if at least one block is selected from the block cache
// but the required launch conditions are not fully satisfied.
// This ensures the system progresses without excessive delays while waiting
// for additional blocks or conditions to meet the launch criteria.
const StrictMaxWaitDelay = 2 * time.Second

// -----------------------------------------------------
