package entity

import (
	"fmt"
	"opl/common"
	"opl/elements"
	"strings"
	"sync"
)

type workerTxPool struct {
	poolMux sync.RWMutex
	txs     map[common.Hash]elements.Transaction
}

func newWorkerTxPool() *workerTxPool {
	return &workerTxPool{
		poolMux: sync.RWMutex{},
		txs:     make(map[common.Hash]elements.Transaction),
	}
}

func (wtp *workerTxPool) containAll(txIDs []common.Hash) bool {
	wtp.poolMux.RLock()
	defer wtp.poolMux.RUnlock()
	for i := 0; i < len(txIDs); i++ {
		if _, ok := wtp.txs[txIDs[i]]; !ok {
			return false
		}
	}
	return true
}

func (wtp *workerTxPool) addTx(tx elements.Transaction) {
	wtp.poolMux.Lock()
	defer wtp.poolMux.Unlock()
	wtp.txs[tx.TxID] = tx
}

func (wtp *workerTxPool) addTxGroup(txs []elements.Transaction) {
	wtp.poolMux.Lock()
	defer wtp.poolMux.Unlock()
	for i := 0; i < len(txs); i++ {
		wtp.txs[txs[i].TxID] = txs[i]
	}
}

// retrieveByIDs retrieve transactions by given IDs.
// NOTE: it must be used after containAll has been called.
func (wtp *workerTxPool) retrieveByIDs(txIDs []common.Hash) []elements.Transaction {
	wtp.poolMux.Lock()
	defer wtp.poolMux.Unlock()
	backTxs := make([]elements.Transaction, len(txIDs))
	for i := 0; i < len(txIDs); i++ {
		if tx, ok := wtp.txs[txIDs[i]]; ok {
			backTxs[i] = tx
			delete(wtp.txs, txIDs[i])
		}
	}
	return backTxs
}

func (wtp *workerTxPool) amount() int {
	wtp.poolMux.RLock()
	defer wtp.poolMux.RUnlock()
	return len(wtp.txs)
}

// retrieveAll retrieve up to maxAmount transactions from the pool.
// back the transactions and their IDs.
func (wtp *workerTxPool) retrieveAll(maxAmount int) ([]elements.Transaction, []common.Hash) {
	wtp.poolMux.Lock()
	defer wtp.poolMux.Unlock()

	var transactions []elements.Transaction
	var txIDs []common.Hash

	for txID, tx := range wtp.txs {
		if len(transactions) >= maxAmount {
			break
		}
		transactions = append(transactions, tx)
		txIDs = append(txIDs, txID)
	}

	for _, txID := range txIDs {
		delete(wtp.txs, txID)
	}
	return transactions, txIDs
}

// String method for workerTxPool
// Returns a formatted string representation of the workerTxPool struct.
func (w *workerTxPool) String() string {
	var builder strings.Builder

	// Lock for reading the transaction pool
	w.poolMux.RLock()
	defer w.poolMux.RUnlock()

	// Print the number of transactions in the pool
	builder.WriteString(fmt.Sprintf("Number of transactions in pool: %d\n", len(w.txs)))

	//// Print the transactions in the pool (if any)
	//if len(w.txs) > 0 {
	//	for txHash, tx := range w.txs {
	//		// Assuming elements.Transaction has a valid String method or you can format it appropriately
	//		builder.WriteString(fmt.Sprintf("Transaction Hash: %v\n", txHash))
	//		builder.WriteString(fmt.Sprintf("Transaction Details: %v\n", tx)) // Replace this with the actual details if necessary
	//	}
	//} else {
	//	builder.WriteString("No transactions in pool.\n")
	//}

	return builder.String()
}
