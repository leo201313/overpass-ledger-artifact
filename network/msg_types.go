package network

// message codes for default connection maintain
const (
	baseOffset = 0xE0

	innerMsg = baseOffset + 0x00 // reserved code for inner network communication
	discMsg  = baseOffset + 0x01
	pingMsg  = baseOffset + 0x02
	pongMsg  = baseOffset + 0x03
)

// message codes for testing the protocol of overpass-ledger
const (
	NormalMsg = 0x00 // reserved code
	TestMsg   = 0x01 // used for test
)
