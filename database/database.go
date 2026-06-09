package database

import "github.com/syndtr/goleveldb/leveldb"

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	Start()
	Flush() error
	BatchWrite(batch *leveldb.Batch) error
}
