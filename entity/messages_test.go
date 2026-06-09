package entity

import (
	"opl/common"
	"opl/rlp"
	"testing"
)

func TestUppPrepareMsg(t *testing.T) {
	prepareMsg := UppPrepareMsg{EpochID: common.GenerateRandomHash()}
	msgBytes, err := rlp.EncodeToBytes(&prepareMsg)
	if err != nil {
		t.Fatal(err)
	}

	var prepareMsg_ UppPrepareMsg
	err = rlp.DecodeBytes(msgBytes, &prepareMsg_)
	if err != nil {
		t.Fatal(err)
	}
	if prepareMsg.EpochID != prepareMsg_.EpochID {
		t.Fatal("Not right")
	}
}
