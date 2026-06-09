package stateManager

import (
	"github.com/syndtr/goleveldb/leveldb"
	"opl/coes"
	"opl/common"
	"opl/database"
	"opl/elements"
	"opl/rlp"
)

// WorldStateManager is the only way to write world state!
type WorldStateManager struct {
	db *database.SimpleLDB
}

func NewWorldStateManager(db *database.SimpleLDB) *WorldStateManager {
	db.Put(coes.EpochHeightBytes, Uint64ToBytes(0))
	return &WorldStateManager{db: db}
}

func (wsm *WorldStateManager) ReadState(addr common.Address) (have bool, value []byte) {
	value, err := wsm.db.Get(addr[:])
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		} else {
			panic(err) // at this version, we do not handle these unexpected errors
		}
	}
	return true, value
}

func (wsm *WorldStateManager) WriteState(addr common.Address, value []byte) {
	if len(value) == 0 { // this state should be deleted
		err := wsm.db.Delete(addr[:])
		if err != nil {
			panic(err) // at this version, we do not handle these unexpected errors
		}
		return
	}

	err := wsm.db.Put(addr[:], value)
	if err != nil {
		panic(err) // at this version, we do not handle these unexpected errors
	}
}

func (wsm *WorldStateManager) CommitStateSet(commitSet []elements.StateCommit) {
	batch := new(leveldb.Batch)
	for _, ws := range commitSet {

		if len(ws.Value) == 0 {
			batch.Delete(ws.Address[:])
		} else {
			batch.Put(ws.Address[:], ws.Value)
		}
	}
	wsm.db.BatchWrite(batch)
}

// AppendEpoch only appnd the epoch on the blockchain and update the height.
// It does not commit the state set on the world states.
func (wsm *WorldStateManager) AppendEpoch(ep elements.Epoch) {
	nowHeightBytes, err := wsm.db.Get(coes.EpochHeightBytes)
	if err != nil {
		panic(err)
	}
	nowHeight, err := BytesToUint64(nowHeightBytes)
	if err != nil {
		panic(err)
	}
	if nowHeight+1 != ep.Height {
		panic("cannot append the epoch into blockchain as height not match")
	}

	nowHeight += 1
	epochBytes, err := rlp.EncodeToBytes(&ep)
	if err != nil {
		panic(err)
	}
	epochKey := EpochToBytes(nowHeight)
	wsm.db.Put(epochKey, epochBytes)
	wsm.db.Put(coes.EpochHeightBytes, Uint64ToBytes(nowHeight))
}

func (wsm *WorldStateManager) CurrentHeight() uint64 {
	nowHeightBytes, err := wsm.db.Get(coes.EpochHeightBytes)
	if err != nil {
		panic(err)
	}
	nowHeight, err := BytesToUint64(nowHeightBytes)
	if err != nil {
		panic(err)
	}
	return nowHeight
}

func (wsm *WorldStateManager) GetEpochByHeight(height uint64) (*elements.Epoch, error) {
	epochKey := EpochToBytes(height)
	epochBytes, err := wsm.db.Get(epochKey)
	if err != nil {
		return nil, err
	}
	var gotEpoch elements.Epoch
	err = rlp.DecodeBytes(epochBytes, &gotEpoch)
	if err != nil {
		return nil, err
	}
	return &gotEpoch, nil
}

func (wsm *WorldStateManager) GetTxIDsInEpochByHeight(height uint64) ([]common.Hash, error) {
	epochKey := EpochToBytes(height)
	epochBytes, err := wsm.db.Get(epochKey)
	if err != nil {
		return nil, err
	}
	var gotEpoch elements.Epoch
	err = rlp.DecodeBytes(epochBytes, &gotEpoch)
	if err != nil {
		return nil, err
	}
	txIDs := gotEpoch.BackAllTxIDs()
	return txIDs, nil
}

func (wsm *WorldStateManager) GetTxIDsAndTypesInEpochByHeight(height uint64) ([]common.Hash, []uint8, error) {
	epochKey := EpochToBytes(height)
	epochBytes, err := wsm.db.Get(epochKey)
	if err != nil {
		return nil, nil, err
	}
	var gotEpoch elements.Epoch
	err = rlp.DecodeBytes(epochBytes, &gotEpoch)
	if err != nil {
		return nil, nil, err
	}
	txIDs, txTypes := gotEpoch.BackAllTxIDsAndHandleType()
	return txIDs, txTypes, nil
}
