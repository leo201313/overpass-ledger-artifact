package smartcontract

import (
	"crypto/sha256"
	"opl/common"
	"opl/elements"
	"sort"
	"strconv"
)

func InvertAsCheckAddr(addr common.Address) common.Address {
	neoAddr := common.Address{}
	copy(neoAddr[:], addr[10:20])
	copy(neoAddr[10:], addr[0:10])
	return neoAddr
}

func BytesToInt(byteSlice []byte) (int, error) {
	return strconv.Atoi(string(byteSlice))
}

func IntToBytes(num int) []byte {
	return []byte(strconv.Itoa(num))
}

// GenerateTestUpdateBalanceTx only used for test
func GenerateTestUpdateBalanceTx(accountName string, amount int) elements.Transaction {
	accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountName)))
	amountBytes := IntToBytes(amount)

	arg0 := elements.Argument{
		Type:    1,
		Address: accountAddr,
		Value:   nil,
	}

	arg1 := elements.Argument{
		Type:    0,
		Address: common.Address{},
		Value:   amountBytes,
	}

	tx := elements.Transaction{
		TxID:          common.Hash{},
		Sender:        common.Address{},
		Version:       common.GenerateRandomHash(),
		Nonce:         0,
		Contract:      DEMO_CONTRACT_SMALLBANK,
		Function:      SMALLBANK_FUNC_UPDATEBALANCE,
		Arguments:     []elements.Argument{arg0, arg1},
		Signature:     nil,
		StateReadSet:  nil,
		StateWriteSet: nil,
		Results:       nil,
	}
	tx.SetTxID()
	return tx
}

// GenerateTestSendPaymentTx only used for test
func GenerateTestSendPaymentTx(accountName1, accountName2 string, amount int) elements.Transaction {
	accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
	accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(accountName2)))
	amountBytes := IntToBytes(amount)

	arg0 := elements.Argument{
		Type:    1,
		Address: accountAddr1,
		Value:   nil,
	}

	arg1 := elements.Argument{
		Type:    1,
		Address: accountAddr2,
		Value:   nil,
	}

	arg2 := elements.Argument{
		Type:    0,
		Address: common.Address{},
		Value:   amountBytes,
	}

	tx := elements.Transaction{
		TxID:          common.Hash{},
		Sender:        common.Address{},
		Version:       common.GenerateRandomHash(),
		Nonce:         0,
		Contract:      DEMO_CONTRACT_SMALLBANK,
		Function:      SMALLBANK_FUNC_SENDPAYMENT,
		Arguments:     []elements.Argument{arg0, arg1, arg2},
		Signature:     nil,
		StateReadSet:  nil,
		StateWriteSet: nil,
		Results:       nil,
	}
	tx.SetTxID()
	return tx
}

func GenerateTestAlmagate(accountName1, accountName2 string) elements.Transaction {
	accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
	accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(accountName2)))

	arg0 := elements.Argument{
		Type:    1,
		Address: accountAddr1,
		Value:   nil,
	}

	arg1 := elements.Argument{
		Type:    1,
		Address: accountAddr2,
		Value:   nil,
	}

	tx := elements.Transaction{
		TxID:          common.Hash{},
		Sender:        common.Address{},
		Version:       common.GenerateRandomHash(),
		Nonce:         0,
		Contract:      DEMO_CONTRACT_SMALLBANK,
		Function:      SMALLBANK_FUNC_ALMAGATE,
		Arguments:     []elements.Argument{arg0, arg1},
		Signature:     nil,
		StateReadSet:  nil,
		StateWriteSet: nil,
		Results:       nil,
	}
	tx.SetTxID()
	return tx
}

func BackAccountNameByIndex(index int) string {
	return "account-" + strconv.Itoa(index)
}

var InitValueBytes = IntToBytes(BANLANCE)

func QSort(arr []int, left int, right int) {
	//i := left
	//j := right
	//
	//if i >= j {
	//	return
	//}
	//
	//pivot := arr[left+(right-left)/2]
	//
	//for i <= j {
	//	for arr[i] < pivot {
	//		i++
	//	}
	//	for arr[j] > pivot {
	//		j--
	//	}
	//	if i <= j {
	//		arr[i], arr[j] = arr[j], arr[i]
	//		i++
	//		j--
	//	}
	//}
	//
	//if left < j {
	//	QSort(arr, left, j)
	//}
	//if i < right {
	//	QSort(arr, i, right)
	//}
	sort.Ints(arr)
}
