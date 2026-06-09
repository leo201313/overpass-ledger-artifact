package utils

import "sort"

func IPAddrCompare(a, b string) bool {
	return a > b
}

func ComputeIndexFromIPAddr(selfAddr string, otherAddr []string) int {
	have := false
	for i := 0; i < len(otherAddr); i++ {
		if otherAddr[i] == selfAddr {
			have = true
			break
		}
	}
	iplist := make([]string, len(otherAddr))
	copy(iplist, otherAddr)
	if !have {
		iplist = append(iplist, selfAddr)
	}
	sort.Slice(iplist, func(i, j int) bool {
		return IPAddrCompare(iplist[i], iplist[j])
	})
	for i := 0; i < len(iplist); i++ {
		if iplist[i] == selfAddr {
			return i
		}
	}
	return 0
}
