package stateManager

import (
	"opl/common"
	"opl/elements"
)

type VersionDependencyTable struct {
	NowBlockID common.Hash
	Prev       map[common.Address]VDTrow
	Now        map[common.Address]VDTrow
}

type VDTrow struct {
	Value   []byte
	Version common.Hash
}

func NewVDT() *VersionDependencyTable {
	return &VersionDependencyTable{
		NowBlockID: common.Hash{},
		Prev:       make(map[common.Address]VDTrow),
		Now:        make(map[common.Address]VDTrow),
	}
}

func (vdt *VersionDependencyTable) ReadState(stateAddr common.Address) (have bool, value []byte, version common.Hash) {
	row1, ok1 := vdt.Now[stateAddr]
	if !ok1 {
		row2, ok2 := vdt.Prev[stateAddr]
		if !ok2 {
			return false, nil, common.Hash{}
		} else {
			return true, row2.Value, row2.Version
		}
	} else {
		return true, row1.Value, row1.Version
	}
}

func (vdt *VersionDependencyTable) WriteState(stateAddr common.Address, value []byte, version common.Hash) {
	row := VDTrow{
		Value:   value,
		Version: version,
	}

	vdt.Now[stateAddr] = row
}

func (vdt *VersionDependencyTable) TrimState(commitState []elements.StateCommit, epochVersion common.Hash) {
	for _, state := range commitState {
		//trim the stable state
		delete(vdt.Now, state.Address)

		//vdt.Now[state.Address] = VDTrow{
		//	Value:   state.Value,
		//	Version: epochVersion,
		//}
	}

	vdt.Prev = vdt.Now
	vdt.Now = make(map[common.Address]VDTrow)
}

func (vdt *VersionDependencyTable) UpdateCurrentBlockID(blockID common.Hash) {
	vdt.NowBlockID = blockID
}
