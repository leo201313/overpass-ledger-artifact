package utils

import "testing"

func TestIPAddrCompare(t *testing.T) {
	addr1 := "127.0.0.1:8030"
	addr2 := "227.3.48.77:8030"
	addr3 := "127.0.0.1:8031"
	addr4 := "8.123.44.132:777"

	t.Log(IPAddrCompare(addr1, addr2))
	t.Log(IPAddrCompare(addr2, addr1))

	t.Log(IPAddrCompare(addr2, addr3))
	t.Log(IPAddrCompare(addr3, addr2))

	t.Log(IPAddrCompare(addr3, addr4))
	t.Log(IPAddrCompare(addr4, addr3))

	t.Log(IPAddrCompare(addr4, addr1))
	t.Log(IPAddrCompare(addr1, addr4))
}

func TestComputeIndexFromIPAddr(t *testing.T) {
	addr1 := "127.0.0.1:8030"
	addr2 := "227.3.48.77:8030"
	addr3 := "127.0.0.1:8031"
	addr4 := "8.123.44.132:777"

	list1 := []string{addr1, addr2, addr3}
	list2 := []string{addr1, addr2, addr3, addr4}

	t.Log(ComputeIndexFromIPAddr(addr4, list1))
	t.Log(ComputeIndexFromIPAddr(addr4, list2))
	t.Log(ComputeIndexFromIPAddr(addr3, list2))
	t.Log(ComputeIndexFromIPAddr(addr2, list2))
	t.Log(ComputeIndexFromIPAddr(addr1, list2))
}
