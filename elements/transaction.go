package elements

import (
	"bytes"
	"crypto/sha256"
	"opl/common"
	"opl/rlp"
)

// Transactions are no longer should be classified into WP and WO
//const (
//	WP_TRANSACTION uint8 = 0
//	WO_TRANSACTION uint8 = 1
//)

const (
	VALUE_ARG uint8 = 0
	ADDR_ARG  uint8 = 1
)

type Transaction struct {
	TxID common.Hash

	Sender  common.Address
	Version common.Hash // the epoch version when published
	Nonce   uint64

	Contract  common.Address
	Function  common.Address
	Arguments []Argument

	Signature []byte

	// After lower layer execution
	StateReadSet  []StateRead
	StateWriteSet []StateWrite
	Results       []byte // to store the temporary results that is back in lower layer
}

type Argument struct {
	Type    uint8 // 0 is plain-value, 1 is the address of state
	Address common.Address
	Value   []byte
}

const (
	EpochVersion uint8 = 0
	BlockVersion uint8 = 1
)

type StateRead struct {
	Address     common.Address
	Version     common.Hash
	VersionType uint8
	Value       []byte
}

type StateWrite struct {
	Address common.Address
	Value   []byte
}

func (tx *Transaction) Hash() common.Hash {
	sender, _ := rlp.EncodeToBytes(tx.Sender)
	version, _ := rlp.EncodeToBytes(tx.Version)
	nonce, _ := rlp.EncodeToBytes(tx.Nonce)
	contract, _ := rlp.EncodeToBytes(tx.Contract)
	function, _ := rlp.EncodeToBytes(tx.Function)
	arguments, _ := rlp.EncodeToBytes(tx.Arguments)

	jointBytes := bytes.Join([][]byte{sender, version, nonce, contract, function, arguments}, []byte{})
	hashBytes := sha256.Sum256(jointBytes)
	return hashBytes
}

func (tx *Transaction) SetTxID() {
	tx.TxID = tx.Hash()
}

func CreateArgument(isAddress bool, address common.Address, val []byte) Argument {
	if isAddress {
		return Argument{
			Type:    1,
			Address: address,
			Value:   nil,
		}
	} else {
		return Argument{
			Type:    0,
			Address: common.Address{},
			Value:   val,
		}
	}
}
