package elements

import (
	"opl/common"
	"testing"
)

func TestTransaction_Hash(t *testing.T) {
	testTx := Transaction{
		TxID:          common.Hash{},
		Sender:        common.BytesToAddress([]byte("test_transaction_sender")),
		Version:       common.BytesToHash([]byte("test_transaction_version")),
		Nonce:         0,
		Contract:      common.BytesToAddress([]byte("test_transaction_sender")),
		Function:      common.BytesToAddress([]byte("test_transaction_sender")),
		Arguments:     []Argument{CreateArgument(false, common.Address{}, []byte("test_transaction_argument"))},
		Signature:     nil,
		StateReadSet:  nil,
		StateWriteSet: nil,
	}
	testTx.SetTxID()
	if testTx.TxID == (common.Hash{}) {
		t.Fatal("The Hash is wrong")
	}
	t.Logf("The TxID after Hash is %x", testTx.TxID)
}
