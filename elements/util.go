package elements

import "opl/common"

func CreateStateRead(addr common.Address, version common.Hash, versionType uint8, value []byte) StateRead {
	return StateRead{
		Address:     addr,
		Version:     version,
		VersionType: versionType,
		Value:       value,
	}
}

func CreateStateWrite(addr common.Address, value []byte) StateWrite {
	return StateWrite{
		Address: addr,
		Value:   value,
	}
}
