package smartcontract

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"opl/common"
	"opl/elements"
)

var (
	DEMO_CONTRACT_KVSTORE = common.HashToAddress(sha256.Sum256([]byte("CONTRACT_KVSTORE")))
	KVSTORE_FUNC_WRITE    = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_WRITE")))
	KVSTORE_FUNC_READ     = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_READ")))
	KVSTORE_FUNC_DELETE   = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_DELETE")))

	DEMO_CONTRACT_SMALLBANK      = common.HashToAddress(sha256.Sum256([]byte("CONTRACT_SMALLBANK")))
	SMALLBANK_FUNC_QUERY         = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_QUERY")))
	SMALLBANK_FUNC_ALMAGATE      = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_ALMAGATE")))
	SMALLBANK_FUNC_GETBALANCE    = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_GETBALANCE")))
	SMALLBANK_FUNC_UPDATEBALANCE = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_UPDATEBALANCE")))
	SMALLBANK_FUNC_UPDATESAVING  = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_UPDATESAVING")))
	SMALLBANK_FUNC_SENDPAYMENT   = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_SENDPAYMENT")))
	SMALLBANK_FUNC_WRITECHECK    = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_WRITECHECK")))

	DEMO_CONTRACT_TRANSFER = common.HashToAddress(sha256.Sum256([]byte("CONTRACT_TRANSFER")))
	TRANSFER_FUNC_TRANSFER = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_TRANSFER")))

	BANLANCE = 100000

	DEMO_CONTRACT_CPUHEAVY = common.HashToAddress(sha256.Sum256([]byte("CONTRACT_CPUHEAVY")))
	CPUHEAVY_FUNC_SORT     = common.HashToAddress(sha256.Sum256([]byte("FUNCTION_SORT")))
)

var (
	ERROR_0 = errors.New("arguments not match")
	ERROR_1 = errors.New("type not match as signed")
)

// DemoSME is a simple smart contract engine that only support
// K-V store and SmallBank two contracts.
// It is only used for test and demo.
type DemoSME struct {
	UppRead  func(addr common.Address) (have bool, value []byte)
	UppWrite func(addr common.Address, value []byte)

	LowRead  func(addr common.Address) (have bool, value []byte, version common.Hash, versionType uint8)
	LowWrite func(addr common.Address, value []byte)

	LocalRead  func(addr common.Address) (have bool, value []byte)
	LocalWrite func(addr common.Address, value []byte)
}

func NewDemoSME() *DemoSME {
	return &DemoSME{}
}

func (sme *DemoSME) UppInstallReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte), writefunc func(addr common.Address, value []byte)) {
	sme.UppRead = readFunc
	sme.UppWrite = writefunc
}
func (sme *DemoSME) UppExecuteTransaction(tx elements.Transaction, preLoaded []elements.StateRead) (result []byte) {

	fastRead := func(addr common.Address) (bool, []byte) {
		for _, rs := range preLoaded {
			if rs.Address == addr {
				if len(rs.Value) == 0 {
					// NOTE: this is important as the lowerStateManger use nil to represent unexisted value
					return false, nil
				}
				return true, rs.Value
			}
		}

		return sme.UppRead(addr)
	}

	switch tx.Contract {
	case DEMO_CONTRACT_TRANSFER:
		switch tx.Function {
		case TRANSFER_FUNC_TRANSFER:
			if len(tx.Arguments) != 3 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 || tx.Arguments[2].Type != 0 {
				panic(ERROR_1)
			}
			fromAddr := tx.Arguments[0].Address
			toAddr := tx.Arguments[1].Address

			amount_bytes := tx.Arguments[2].Value

			have, bal_bytes1 := fastRead(fromAddr)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}
			have, bal_bytes2 := fastRead(toAddr)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			amount, err := BytesToInt(amount_bytes)
			if err != nil {
				panic(err)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(bal_bytes1), err))
			}
			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}
			bal1 -= amount
			bal2 += amount
			sme.UppWrite(fromAddr, IntToBytes(bal1))
			sme.UppWrite(toAddr, IntToBytes(bal2))
			return nil

		default:
			panic("unknown function found in CONTRACT_TRANSFER")

		}

	case DEMO_CONTRACT_CPUHEAVY:
		switch tx.Function {
		case CPUHEAVY_FUNC_SORT:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 0 {
				panic(ERROR_1)
			}

			sizeBytes := tx.Arguments[0].Value
			size, err := BytesToInt(sizeBytes)
			if err != nil {
				panic(err)
			}
			arr := make([]int, size)
			for i := 0; i < size; i++ {
				arr[i] = size - i
			}
			QSort(arr, 0, size-1)
			return []byte("Succeed")

		default:
			panic("unknown function found in CONTRACT_CPUHEAVY")
		}

	case DEMO_CONTRACT_KVSTORE:
		switch tx.Function {
		case KVSTORE_FUNC_WRITE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			sme.UppWrite(tx.Arguments[0].Address, tx.Arguments[1].Value)
			return nil

		case KVSTORE_FUNC_READ:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			have, value := fastRead(tx.Arguments[0].Address)
			if have {
				return value
			} else {
				return []byte("")
			}

		case KVSTORE_FUNC_DELETE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			sme.UppWrite(tx.Arguments[0].Address, nil)
			return nil

		default:
			panic("unknown function found in CONTRACT_KVSTORE")
		}
	case DEMO_CONTRACT_SMALLBANK:
		switch tx.Function {
		case SMALLBANK_FUNC_QUERY:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			have, value := fastRead(tx.Arguments[0].Address)
			if have {
				return value
			} else {
				return []byte("")
			}
		case SMALLBANK_FUNC_ALMAGATE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr_From := tx.Arguments[0].Address
			savingAddr_To := tx.Arguments[1].Address
			checkAddr_From := InvertAsCheckAddr(savingAddr_From)
			checkAddr_To := InvertAsCheckAddr(savingAddr_To)

			have, val1 := fastRead(savingAddr_From)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}

			have, val2 := fastRead(checkAddr_To)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(val1), err))
			}

			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.UppWrite(checkAddr_From, IntToBytes(0))
			sme.UppWrite(savingAddr_To, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_GETBALANCE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, val1 := fastRead(savingAddr)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}
			have, val2 := fastRead(checkAddr)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(val1), err))
			}
			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}
			bal1 += bal2
			return IntToBytes(bal1)

		case SMALLBANK_FUNC_UPDATEBALANCE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal1_bytes := fastRead(checkAddr)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(bal1_bytes), err))
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.UppWrite(checkAddr, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_UPDATESAVING:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address

			have, bal1_bytes := fastRead(savingAddr)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(bal1_bytes), err))
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.UppWrite(savingAddr, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_SENDPAYMENT:
			if len(tx.Arguments) != 3 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 || tx.Arguments[2].Type != 0 {
				panic(ERROR_1)
			}
			fromAddr := tx.Arguments[0].Address
			fromCheckAddr := InvertAsCheckAddr(fromAddr)
			toAddr := tx.Arguments[1].Address
			toCheckAddr := InvertAsCheckAddr(toAddr)

			amount_bytes := tx.Arguments[2].Value

			have, bal_bytes1 := fastRead(fromCheckAddr)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}
			have, bal_bytes2 := fastRead(toCheckAddr)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			amount, err := BytesToInt(amount_bytes)
			if err != nil {
				panic(err)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(bal_bytes1), err))
			}
			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}
			bal1 -= amount
			bal2 += amount
			sme.UppWrite(fromCheckAddr, IntToBytes(bal1))
			sme.UppWrite(toCheckAddr, IntToBytes(bal2))

			return nil

		case SMALLBANK_FUNC_WRITECHECK:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal_bytes1 := fastRead(checkAddr)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}
			have, bal_bytes2 := fastRead(savingAddr)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(fmt.Errorf("the bytes len(val1) is %d, %v", len(bal_bytes1), err))
			}

			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}

			amount, err := BytesToInt(tx.Arguments[1].Value)
			if err != nil {
				panic(err)
			}

			if amount < (bal1 + bal2) {
				sme.UppWrite(checkAddr, IntToBytes(bal1-amount-1))
			} else {
				sme.UppWrite(checkAddr, IntToBytes(bal1-amount))
			}
			return nil
		default:
			panic("unknown function found in CONTRACT_SMALLBANK")

		}
	default:
		panic("unknown contract is found")
	}
	return nil

}
func (sme *DemoSME) LowInstallReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte, version common.Hash, versionType uint8), writefunc func(addr common.Address, value []byte)) {
	sme.LowRead = readFunc
	sme.LowWrite = writefunc
}

func (sme *DemoSME) LowExecuteTransaction(tx elements.Transaction) (readSet []elements.StateRead, writeSet []elements.StateWrite, result []byte) {
	switch tx.Contract {
	case DEMO_CONTRACT_TRANSFER:
		switch tx.Function {
		case TRANSFER_FUNC_TRANSFER:
			if len(tx.Arguments) != 3 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 || tx.Arguments[2].Type != 0 {
				panic(ERROR_1)
			}
			fromAddr := tx.Arguments[0].Address

			toAddr := tx.Arguments[1].Address

			amount_bytes := tx.Arguments[2].Value

			have, bal_bytes1, version, versionType := sme.LowRead(fromAddr)
			rs := elements.CreateStateRead(fromAddr, version, versionType, bal_bytes1)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}

			have, bal_bytes2, version, versionType := sme.LowRead(toAddr)
			rs = elements.CreateStateRead(toAddr, version, versionType, bal_bytes2)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			amount, err := BytesToInt(amount_bytes)
			if err != nil {
				panic(err)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}
			bal1 -= amount
			bal2 += amount

			writeBytes1 := IntToBytes(bal1)
			writeBytes2 := IntToBytes(bal2)

			sme.LowWrite(fromAddr, writeBytes1)
			ws := elements.CreateStateWrite(fromAddr, writeBytes1)
			writeSet = append(writeSet, ws)

			sme.LowWrite(toAddr, writeBytes2)
			ws = elements.CreateStateWrite(toAddr, writeBytes2)
			writeSet = append(writeSet, ws)

			return readSet, writeSet, result

		default:
			panic("unknown function found in CONTRACT_TRANSFER")
		}

	case DEMO_CONTRACT_CPUHEAVY:
		switch tx.Function {
		case CPUHEAVY_FUNC_SORT:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 0 {
				panic(ERROR_1)
			}

			sizeBytes := tx.Arguments[0].Value
			size, err := BytesToInt(sizeBytes)
			if err != nil {
				panic(err)
			}
			arr := make([]int, size)
			for i := 0; i < size; i++ {
				arr[i] = size - i
			}
			QSort(arr, 0, size-1)
			return nil, nil, []byte("Succeed")

		default:
			panic("unknown function found in CONTRACT_CPUHEAVY")
		}

	case DEMO_CONTRACT_KVSTORE:
		switch tx.Function {
		case KVSTORE_FUNC_READ:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			have, value, version, versionType := sme.LowRead(tx.Arguments[0].Address)

			rs := elements.StateRead{
				Address:     tx.Arguments[0].Address,
				Version:     version,
				VersionType: versionType,
				Value:       value,
			}

			readSet = append(readSet, rs)

			if have {
				result = value
			} else {
				result = []byte("")
			}

			return readSet, writeSet, result

		case KVSTORE_FUNC_WRITE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}

			sw := elements.StateWrite{
				Address: tx.Arguments[0].Address,
				Value:   tx.Arguments[1].Value,
			}
			writeSet = append(writeSet, sw)
			sme.LowWrite(tx.Arguments[0].Address, tx.Arguments[1].Value)
			return readSet, writeSet, result
		case KVSTORE_FUNC_DELETE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			sw := elements.StateWrite{
				Address: tx.Arguments[0].Address,
				Value:   nil,
			}
			writeSet = append(writeSet, sw)
			sme.LowWrite(tx.Arguments[0].Address, nil)
			return readSet, writeSet, result
		default:
			panic("unknown function found in CONTRACT_KVSTORE")

		}
	case DEMO_CONTRACT_SMALLBANK:
		switch tx.Function {
		case SMALLBANK_FUNC_QUERY:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			have, value, version, versionType := sme.LowRead(tx.Arguments[0].Address)
			rs := elements.CreateStateRead(tx.Arguments[0].Address, version, versionType, value)
			readSet = append(readSet, rs)
			if have {
				result = value
			} else {
				result = []byte("")
			}
			return readSet, writeSet, result

		case SMALLBANK_FUNC_ALMAGATE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr_From := tx.Arguments[0].Address
			savingAddr_To := tx.Arguments[1].Address
			checkAddr_From := InvertAsCheckAddr(savingAddr_From)
			checkAddr_To := InvertAsCheckAddr(savingAddr_To)

			have, val1, version, versionType := sme.LowRead(savingAddr_From)
			rs := elements.CreateStateRead(savingAddr_From, version, versionType, val1)
			readSet = append(readSet, rs)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}

			have, val2, version, versionType := sme.LowRead(checkAddr_To)
			rs = elements.CreateStateRead(checkAddr_To, version, versionType, val2)
			readSet = append(readSet, rs)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(err)
			}

			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}

			bal1 += bal2

			sme.LowWrite(checkAddr_From, IntToBytes(0))
			ws := elements.CreateStateWrite(checkAddr_From, IntToBytes(0))
			writeSet = append(writeSet, ws)

			sme.LowWrite(savingAddr_To, IntToBytes(bal1))
			ws = elements.CreateStateWrite(savingAddr_To, IntToBytes(bal1))
			writeSet = append(writeSet, ws)

			return readSet, writeSet, result

		case SMALLBANK_FUNC_GETBALANCE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, val1, version, versionType := sme.LowRead(savingAddr)
			rs := elements.CreateStateRead(savingAddr, version, versionType, val1)
			readSet = append(readSet, rs)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}

			have, val2, version, versionType := sme.LowRead(checkAddr)
			rs = elements.CreateStateRead(checkAddr, version, versionType, val2)
			readSet = append(readSet, rs)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}
			bal1 += bal2

			result = IntToBytes(bal1)
			return readSet, writeSet, result

		case SMALLBANK_FUNC_UPDATEBALANCE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal1_bytes, version, versionType := sme.LowRead(checkAddr)
			rs := elements.CreateStateRead(checkAddr, version, versionType, bal1_bytes)
			readSet = append(readSet, rs)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			writeBytes := IntToBytes(bal1)
			sme.LowWrite(checkAddr, writeBytes)
			ws := elements.CreateStateWrite(checkAddr, writeBytes)
			writeSet = append(writeSet, ws)
			return readSet, writeSet, result

		case SMALLBANK_FUNC_UPDATESAVING:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address

			have, bal1_bytes, version, versionType := sme.LowRead(savingAddr)
			rs := elements.CreateStateRead(savingAddr, version, versionType, bal1_bytes)
			readSet = append(readSet, rs)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2

			writeBytes := IntToBytes(bal1)
			sme.LowWrite(savingAddr, writeBytes)
			ws := elements.CreateStateWrite(savingAddr, writeBytes)
			writeSet = append(writeSet, ws)

			return readSet, writeSet, result

		case SMALLBANK_FUNC_SENDPAYMENT:
			if len(tx.Arguments) != 3 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 || tx.Arguments[2].Type != 0 {
				panic(ERROR_1)
			}
			fromAddr := tx.Arguments[0].Address
			fromCheckAddr := InvertAsCheckAddr(fromAddr)
			toAddr := tx.Arguments[1].Address
			toCheckAddr := InvertAsCheckAddr(toAddr)

			amount_bytes := tx.Arguments[2].Value

			have, bal_bytes1, version, versionType := sme.LowRead(fromCheckAddr)
			rs := elements.CreateStateRead(fromCheckAddr, version, versionType, bal_bytes1)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}

			have, bal_bytes2, version, versionType := sme.LowRead(toCheckAddr)
			rs = elements.CreateStateRead(toCheckAddr, version, versionType, bal_bytes2)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			amount, err := BytesToInt(amount_bytes)
			if err != nil {
				panic(err)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}
			bal1 -= amount
			bal2 += amount

			writeBytes1 := IntToBytes(bal1)
			writeBytes2 := IntToBytes(bal2)

			sme.LowWrite(fromCheckAddr, writeBytes1)
			ws := elements.CreateStateWrite(fromCheckAddr, writeBytes1)
			writeSet = append(writeSet, ws)

			sme.LowWrite(toCheckAddr, writeBytes2)
			ws = elements.CreateStateWrite(toCheckAddr, writeBytes2)
			writeSet = append(writeSet, ws)

			return readSet, writeSet, result

		case SMALLBANK_FUNC_WRITECHECK:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal_bytes1, version, versionType := sme.LowRead(checkAddr)
			rs := elements.CreateStateRead(checkAddr, version, versionType, bal_bytes1)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}

			have, bal_bytes2, version, versionType := sme.LowRead(savingAddr)
			rs = elements.CreateStateRead(savingAddr, version, versionType, bal_bytes2)
			readSet = append(readSet, rs)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(err)
			}

			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}

			amount, err := BytesToInt(tx.Arguments[1].Value)
			if err != nil {
				panic(err)
			}

			if amount < (bal1 + bal2) {
				writeBytes := IntToBytes(bal1 - amount - 1)
				sme.LowWrite(checkAddr, writeBytes)
				ws := elements.CreateStateWrite(checkAddr, writeBytes)
				writeSet = append(writeSet, ws)
			} else {
				writeBytes := IntToBytes(bal1 - amount)
				sme.LowWrite(checkAddr, writeBytes)
				ws := elements.CreateStateWrite(checkAddr, writeBytes)
				writeSet = append(writeSet, ws)
			}

			return readSet, writeSet, result

		default:
			panic("unknown function found in CONTRACT_SMALLBANK")
		}
	default:
		panic("unknown contract is found")
	}
}

func (sme *DemoSME) InstallLocalUsedReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte), writefunc func(addr common.Address, value []byte)) {
	sme.LocalRead = readFunc
	sme.LocalWrite = writefunc
}

func (sme *DemoSME) LocalExecuteTransaction(tx elements.Transaction) (result []byte) {
	switch tx.Contract {
	case DEMO_CONTRACT_KVSTORE:
		switch tx.Function {
		case KVSTORE_FUNC_WRITE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			sme.LocalWrite(tx.Arguments[0].Address, tx.Arguments[1].Value)
			return nil

		case KVSTORE_FUNC_READ:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			have, value := sme.LocalRead(tx.Arguments[0].Address)
			if have {
				return value
			} else {
				return []byte("")
			}

		case KVSTORE_FUNC_DELETE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			sme.LocalWrite(tx.Arguments[0].Address, nil)
			return nil

		default:
			panic("unknown function found in CONTRACT_KVSTORE")
		}
	case DEMO_CONTRACT_SMALLBANK:
		switch tx.Function {
		case SMALLBANK_FUNC_QUERY:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}
			have, value := sme.LocalRead(tx.Arguments[0].Address)
			if have {
				return value
			} else {
				return []byte("")
			}
		case SMALLBANK_FUNC_ALMAGATE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr_From := tx.Arguments[0].Address
			savingAddr_To := tx.Arguments[1].Address
			checkAddr_From := InvertAsCheckAddr(savingAddr_From)
			checkAddr_To := InvertAsCheckAddr(savingAddr_To)

			have, val1 := sme.LocalRead(savingAddr_From)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}

			have, val2 := sme.LocalRead(checkAddr_To)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(err)
			}

			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.LocalWrite(checkAddr_From, IntToBytes(0))
			sme.LocalWrite(savingAddr_To, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_GETBALANCE:
			if len(tx.Arguments) != 1 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, val1 := sme.LocalRead(savingAddr)
			if !have {
				val1 = IntToBytes(BANLANCE)
			}
			have, val2 := sme.LocalRead(checkAddr)
			if !have {
				val2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(val1)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(val2)
			if err != nil {
				panic(err)
			}
			bal1 += bal2
			return IntToBytes(bal1)

		case SMALLBANK_FUNC_UPDATEBALANCE:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal1_bytes := sme.LocalRead(checkAddr)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.LocalWrite(checkAddr, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_UPDATESAVING:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}
			savingAddr := tx.Arguments[0].Address

			have, bal1_bytes := sme.LocalRead(savingAddr)
			if !have {
				bal1_bytes = IntToBytes(BANLANCE)
			}

			bal2_bytes := tx.Arguments[1].Value

			bal1, err := BytesToInt(bal1_bytes)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal2_bytes)
			if err != nil {
				panic(err)
			}

			bal1 += bal2
			sme.LocalWrite(savingAddr, IntToBytes(bal1))
			return nil

		case SMALLBANK_FUNC_SENDPAYMENT:
			if len(tx.Arguments) != 3 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 1 || tx.Arguments[2].Type != 0 {
				panic(ERROR_1)
			}
			fromAddr := tx.Arguments[0].Address
			fromCheckAddr := InvertAsCheckAddr(fromAddr)
			toAddr := tx.Arguments[1].Address
			toCheckAddr := InvertAsCheckAddr(toAddr)

			amount_bytes := tx.Arguments[2].Value

			have, bal_bytes1 := sme.LocalRead(fromCheckAddr)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}
			have, bal_bytes2 := sme.LocalRead(toCheckAddr)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			amount, err := BytesToInt(amount_bytes)
			if err != nil {
				panic(err)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(err)
			}
			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}
			bal1 -= amount
			bal2 += amount
			sme.LocalWrite(fromCheckAddr, IntToBytes(bal1))
			sme.LocalWrite(toCheckAddr, IntToBytes(bal2))

			return nil

		case SMALLBANK_FUNC_WRITECHECK:
			if len(tx.Arguments) != 2 {
				panic(ERROR_0)
			}
			if tx.Arguments[0].Type != 1 || tx.Arguments[1].Type != 0 {
				panic(ERROR_1)
			}

			savingAddr := tx.Arguments[0].Address
			checkAddr := InvertAsCheckAddr(savingAddr)

			have, bal_bytes1 := sme.LocalRead(checkAddr)
			if !have {
				bal_bytes1 = IntToBytes(BANLANCE)
			}
			have, bal_bytes2 := sme.LocalRead(savingAddr)
			if !have {
				bal_bytes2 = IntToBytes(BANLANCE)
			}

			bal1, err := BytesToInt(bal_bytes1)
			if err != nil {
				panic(err)
			}

			bal2, err := BytesToInt(bal_bytes2)
			if err != nil {
				panic(err)
			}

			amount, err := BytesToInt(tx.Arguments[1].Value)
			if err != nil {
				panic(err)
			}

			if amount < (bal1 + bal2) {
				sme.LocalWrite(checkAddr, IntToBytes(bal1-amount-1))
			} else {
				sme.LocalWrite(checkAddr, IntToBytes(bal1-amount))
			}
			return nil
		default:
			panic("unknown function found in CONTRACT_SMALLBANK")

		}
	default:
		panic("unknown contract is found")
	}
	return nil
}
