package network

import "opl/utils"

type DialTask struct {
	remoteAddr string
	shouldEnc  bool
	maxTry     int
}

func SimpleCreateDialTasks(selfAddr string, otherAddrs []string, maxTry int, shouldEnc bool) []DialTask {
	tasks := make([]DialTask, 0)
	for _, remoteAddr := range otherAddrs {
		if utils.IPAddrCompare(selfAddr, remoteAddr) { // means should be dialer
			task := DialTask{
				remoteAddr: remoteAddr,
				shouldEnc:  shouldEnc,
				maxTry:     maxTry,
			}
			tasks = append(tasks, task)
		} else { // be receiver and do nothing
			continue
		}
	}
	return tasks
}

func SimpleCreateDialTasks_All(selfAddr string, otherAddrs []string, maxTry int, shouldEnc bool) []DialTask {
	tasks := make([]DialTask, 0)
	for _, remoteAddr := range otherAddrs {
		task := DialTask{
			remoteAddr: remoteAddr,
			shouldEnc:  shouldEnc,
			maxTry:     maxTry,
		}
		tasks = append(tasks, task)
	}
	return tasks
}
