package stateManager

import (
	"opl/common"
	"opl/elements"
	"sort"
)

type VersionCommitmentTable struct {
	BaseVersion common.Hash
	Content     map[common.Address]VCTrow
}

type VCTrow struct {
	Value       []byte
	Version     common.Hash
	VersionType uint8
}

func NewVCT(baseVersion common.Hash) *VersionCommitmentTable {
	return &VersionCommitmentTable{
		BaseVersion: baseVersion,
		Content:     make(map[common.Address]VCTrow),
	}
}

func (vct *VersionCommitmentTable) ReadState(stateAddr common.Address) (bool, []byte, common.Hash, uint8) {
	row, ok := vct.Content[stateAddr]
	if !ok {
		return false, nil, common.Hash{}, 0
	}
	return true, row.Value, row.Version, row.VersionType
}

func (vct *VersionCommitmentTable) WriteState(stateAddr common.Address, value []byte, version common.Hash, versionType uint8) {
	tempRow := VCTrow{
		Value:       value,
		Version:     version,
		VersionType: versionType,
	}
	vct.Content[stateAddr] = tempRow
}

func (vct *VersionCommitmentTable) ExportStates() []elements.StateCommit {
	states := make([]elements.StateCommit, len(vct.Content))
	index := 0
	for addr, row := range vct.Content {
		tempStateCommit := elements.StateCommit{
			Version: row.Version,
			Address: addr,
			Value:   row.Value,
		}
		states[index] = tempStateCommit
		index += 1
	}

	sort.Slice(states, func(i, j int) bool {
		return common.AddressLess(states[i].Address, states[j].Address)
	})

	return states
}
