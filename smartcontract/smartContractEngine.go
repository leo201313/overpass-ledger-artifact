package smartcontract

import (
	"opl/common"
	"opl/elements"
)

type SmartContractEngine interface {
	UppInstallReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte), writefunc func(addr common.Address, value []byte))
	UppExecuteTransaction(tx elements.Transaction, preLoaded []elements.StateRead) (result []byte)
	LowInstallReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte, version common.Hash, versionType uint8), writefunc func(addr common.Address, value []byte))
	LowExecuteTransaction(tx elements.Transaction) (readSet []elements.StateRead, writeSet []elements.StateWrite, result []byte)
	InstallLocalUsedReadWriteFunc(readFunc func(addr common.Address) (have bool, value []byte), writefunc func(addr common.Address, value []byte))
	LocalExecuteTransaction(tx elements.Transaction) (result []byte)
}
