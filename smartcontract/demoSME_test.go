package smartcontract

import (
	"crypto/sha256"
	"opl/common"
	"testing"
)

func TestDemoSME_UppExecuteTransaction(t *testing.T) {
	testAddr := common.HashToAddress(sha256.Sum256([]byte("For Test Use")))
	invertAddr := InvertAsCheckAddr(testAddr)
	t.Logf("the original: %x", testAddr)
	t.Logf("the inverted: %x", invertAddr)
}
