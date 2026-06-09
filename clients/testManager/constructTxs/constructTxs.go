package constructTxs

import (
	"crypto/sha256"
	rand "math/rand/v2"
	"opl/common"
	"opl/elements"
	"opl/smartcontract"
)

func GenerateTransactions_SmallBank(amount int, accountsAmount int) []elements.Transaction {
	txs := make([]elements.Transaction, amount)
	for i := 0; i < amount; i++ {
		r0 := rand.Int() % 6
		switch r0 {
		case 0: //ALMAGATE
			r1 := rand.Int() % accountsAmount
			r2 := rand.Int() % accountsAmount

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
			accountName1 := smartcontract.BackAccountNameByIndex(r1)
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
			accountName1 := smartcontract.BackAccountNameByIndex(r1)
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
			accountName1 := smartcontract.BackAccountNameByIndex(r1)
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
				Contract:  smartcontract.DEMO_CONTRACT_SMALLBANK,
				Function:  smartcontract.SMALLBANK_FUNC_SENDPAYMENT,
				Arguments: []elements.Argument{arg1, arg2, arg3},
				Signature: nil,
			}
			tx.SetTxID()
			txs[i] = tx

		case 5: //WRITECHECK
			r1 := rand.Int() % accountsAmount
			accountName1 := smartcontract.BackAccountNameByIndex(r1)
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

var index = 0

func GenerateTransactions_Transfer(amount int, accountsAmount int) []elements.Transaction {
	txs := make([]elements.Transaction, amount)

	mid := accountsAmount / 2

	for i := 0; i < amount; i++ {
		r1 := index % accountsAmount
		r2 := (index + mid) % accountsAmount
		//r1 := rand.Int() % accountsAmount
		//r2 := rand.Int() % accountsAmount

		index += 1

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

func GenerateTransactions_CPUHeavy(amount int, size int) []elements.Transaction {
	txs := make([]elements.Transaction, amount)
	for i := 0; i < amount; i++ {
		arg0 := elements.Argument{
			Type:    0,
			Address: common.Address{},
			Value:   smartcontract.IntToBytes(size),
		}

		senderAddr := common.HashToAddress(common.GenerateRandomHash())

		tx := elements.Transaction{
			TxID:      common.Hash{},
			Sender:    senderAddr,
			Version:   common.INITIAL_VERSION,
			Nonce:     uint64(i),
			Contract:  smartcontract.DEMO_CONTRACT_CPUHEAVY,
			Function:  smartcontract.CPUHEAVY_FUNC_SORT,
			Arguments: []elements.Argument{arg0},
			Signature: nil,
		}
		tx.SetTxID()
		txs[i] = tx
	}
	return txs
}
