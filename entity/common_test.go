package entity

import (
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
	"opl/common"
	"opl/elements"
	"opl/smartcontract"
	"strconv"
)

var cdntAddrs_test = []string{"127.0.0.1:3030", "127.0.0.1:3031", "127.0.0.1:3032", "127.0.0.1:3033"}
var workerAddrs_test = []string{"127.0.0.1:4030", "127.0.0.1:4031", "127.0.0.1:4032", "127.0.0.1:4033"}
var parties_test = []string{"Org1", "Org2", "Org3", "Org4"}

var shardNO_test uint8 = 1

func generateTxKVStore(shardNO uint8, amount int) []elements.Transaction {

	res := make([]elements.Transaction, amount)
	for i := 0; i < amount; i++ {
		r1 := rand.Int() % 3
		switch r1 {
		case 0: // read
			r2 := rand.Int() % 100
			accountNmae := fmt.Sprintf("Shard-%d-%d", shardNO, r2)
			accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountNmae)))

			arg1 := elements.Argument{
				Type:    elements.ADDR_ARG,
				Address: accountAddr,
				Value:   nil,
			}

			tx := elements.Transaction{
				TxID:          common.Hash{},
				Sender:        common.Address{},
				Version:       common.GenerateRandomHash(),
				Nonce:         0,
				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
				Function:      smartcontract.KVSTORE_FUNC_READ,
				Arguments:     []elements.Argument{arg1},
				Signature:     nil,
				StateReadSet:  nil,
				StateWriteSet: nil,
				Results:       nil,
			}

			tx.SetTxID()
			res[i] = tx

		case 1: // write
			r2 := rand.Int() % 100
			accountNmae := fmt.Sprintf("Shard-%d-%d", shardNO, r2)
			accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountNmae)))
			arg1 := elements.Argument{
				Type:    elements.ADDR_ARG,
				Address: accountAddr,
				Value:   nil,
			}

			r3 := rand.Int() % 100000
			arg2 := elements.Argument{
				Type:    elements.VALUE_ARG,
				Address: common.Address{},
				Value:   smartcontract.IntToBytes(r3),
			}

			tx := elements.Transaction{
				TxID:          common.Hash{},
				Sender:        common.Address{},
				Version:       common.GenerateRandomHash(),
				Nonce:         0,
				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
				Function:      smartcontract.KVSTORE_FUNC_WRITE,
				Arguments:     []elements.Argument{arg1, arg2},
				Signature:     nil,
				StateReadSet:  nil,
				StateWriteSet: nil,
				Results:       nil,
			}

			tx.SetTxID()
			res[i] = tx

		case 2: //delete
			r2 := rand.Int() % 100
			accountNmae := fmt.Sprintf("Shard-%d-%d", shardNO, r2)
			accountAddr := common.HashToAddress(sha256.Sum256([]byte(accountNmae)))

			arg1 := elements.Argument{
				Type:    elements.ADDR_ARG,
				Address: accountAddr,
				Value:   nil,
			}

			tx := elements.Transaction{
				TxID:          common.Hash{},
				Sender:        common.Address{},
				Version:       common.GenerateRandomHash(),
				Nonce:         0,
				Contract:      smartcontract.DEMO_CONTRACT_KVSTORE,
				Function:      smartcontract.KVSTORE_FUNC_DELETE,
				Arguments:     []elements.Argument{arg1},
				Signature:     nil,
				StateReadSet:  nil,
				StateWriteSet: nil,
				Results:       nil,
			}

			tx.SetTxID()
			res[i] = tx
		}

	}
	return res
}

func generateSmallBankTxsTransfer(fromName, toName string, total int) []elements.Transaction {
	res := make([]elements.Transaction, total)
	for i := 0; i < total; i++ {

		accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(fromName)))
		accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(toName)))

		arg1 := elements.Argument{
			Type:    1,
			Address: accountAddr1,
			Value:   nil,
		}

		arg2 := elements.Argument{
			Type:    1,
			Address: accountAddr2,
			Value:   nil,
		}

		arg3 := elements.Argument{
			Type:    0,
			Address: common.Address{},
			Value:   smartcontract.IntToBytes(1),
		}

		senderAddr := common.HashToAddress(common.GenerateRandomHash())

		r1 := uint64(rand.Int())

		tx := elements.Transaction{
			TxID:          common.Hash{},
			Sender:        senderAddr,
			Version:       common.Hash{},
			Nonce:         r1,
			Contract:      smartcontract.DEMO_CONTRACT_SMALLBANK,
			Function:      smartcontract.SMALLBANK_FUNC_SENDPAYMENT,
			Arguments:     []elements.Argument{arg1, arg2, arg3},
			Signature:     nil,
			StateReadSet:  nil,
			StateWriteSet: nil,
			Results:       nil,
		}
		tx.SetTxID()
		res[i] = tx
	}
	return res
}

var accountsAmount = 10000

func backAccountName(index int) string {
	return "account-" + strconv.Itoa(index)
}

func generateTransactions_SmallBank(amount int) []elements.Transaction {
	txs := make([]elements.Transaction, amount)
	for i := 0; i < amount; i++ {
		r0 := rand.Int() % 6
		switch r0 {
		case 0: //ALMAGATE
			r1 := rand.Int() % accountsAmount
			r2 := rand.Int() % accountsAmount

			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))

			accountName2 := backAccountName(r2)
			accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(accountName2)))

			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}

			arg2 := elements.Argument{
				Type:    1,
				Address: accountAddr2,
				Value:   nil,
			}

			senderAddr := common.HashToAddress(common.GenerateRandomHash())

			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_ALMAGATE,
				Arguments: []elements.Argument{arg1, arg2},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx
		case 1: //GETBALANCE
			r1 := rand.Int() % accountsAmount
			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}
			senderAddr := common.HashToAddress(common.GenerateRandomHash())
			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_GETBALANCE,
				Arguments: []elements.Argument{arg1},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx
		case 2: // UPDATEBALANCE
			r1 := rand.Int() % accountsAmount
			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}
			arg2 := elements.Argument{
				Type:    0,
				Address: common.Address{},
				Value:   smartcontract.IntToBytes(smartcontract.BANLANCE),
			}
			senderAddr := common.HashToAddress(common.GenerateRandomHash())
			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_UPDATEBALANCE,
				Arguments: []elements.Argument{arg1, arg2},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx

		case 3: // UPDATESAVING
			r1 := rand.Int() % accountsAmount
			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}
			arg2 := elements.Argument{
				Type:    0,
				Address: common.Address{},
				Value:   smartcontract.IntToBytes(smartcontract.BANLANCE),
			}
			senderAddr := common.HashToAddress(common.GenerateRandomHash())
			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_UPDATESAVING,
				Arguments: []elements.Argument{arg1, arg2},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx

		case 4: // SENDPAYMENT

			r1 := rand.Int() % accountsAmount
			r2 := rand.Int() % accountsAmount

			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))

			accountName2 := backAccountName(r2)
			accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(accountName2)))

			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}

			arg2 := elements.Argument{
				Type:    1,
				Address: accountAddr2,
				Value:   nil,
			}

			arg3 := elements.Argument{
				Type:    0,
				Address: common.Address{},
				Value:   smartcontract.IntToBytes(1),
			}

			senderAddr := common.HashToAddress(common.GenerateRandomHash())

			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_SENDPAYMENT,
				Arguments: []elements.Argument{arg1, arg2, arg3},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx

		case 5: //WRITECHECK
			r1 := rand.Int() % accountsAmount
			accountName1 := backAccountName(r1)
			accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))
			arg1 := elements.Argument{
				Type:    1,
				Address: accountAddr1,
				Value:   nil,
			}
			arg2 := elements.Argument{
				Type:    0,
				Address: common.Address{},
				Value:   smartcontract.IntToBytes(smartcontract.BANLANCE),
			}
			senderAddr := common.HashToAddress(common.GenerateRandomHash())
			tx := elements.Transaction{
				TxID:      common.Hash{},
				Sender:    senderAddr,
				Version:   common.INITIAL_VERSION,
				Nonce:     0,
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_WRITECHECK,
				Arguments: []elements.Argument{arg1, arg2},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx

		}
	}
	return txs
}

func generateTransactions_Transfer(startIndex, amount int, accountsAmount int) []elements.Transaction {
	txs := make([]elements.Transaction, amount)

	mid := accountsAmount / 2

	for i := 0; i < amount; i++ {
		r1 := startIndex % accountsAmount
		r2 := (startIndex + mid) % accountsAmount
		startIndex += 1

		accountName1 := smartcontract.BackAccountNameByIndex(r1)
		accountAddr1 := common.HashToAddress(sha256.Sum256([]byte(accountName1)))

		accountName2 := smartcontract.BackAccountNameByIndex(r2)
		accountAddr2 := common.HashToAddress(sha256.Sum256([]byte(accountName2)))

		arg1 := elements.Argument{
			Type:    1,
			Address: accountAddr1,
			Value:   nil,
		}

		arg2 := elements.Argument{
			Type:    1,
			Address: accountAddr2,
			Value:   nil,
		}

		arg3 := elements.Argument{
			Type:    0,
			Address: common.Address{},
			Value:   smartcontract.IntToBytes(1),
		}

		senderAddr := common.HashToAddress(common.GenerateRandomHash())

		tx := elements.Transaction{
			TxID:      common.Hash{},
			Sender:    senderAddr,
			Version:   common.INITIAL_VERSION,
			Nonce:     0,
			Contract:  smartcontract.DEMO_CONTRACT_TRANSFER,
			Function:  smartcontract.TRANSFER_FUNC_TRANSFER,
			Arguments: []elements.Argument{arg1, arg2, arg3},
			Signature: nil,
		}
		tx.SetTxID()
		txs[i] = tx
	}

	return txs
}
