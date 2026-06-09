package network

import (
	"fmt"
	"opl/rlp"
	"strconv"
	"testing"
	"time"
)

var (
	testAddress = []string{"127.0.0.1:3030", "127.0.0.1:3031", "127.0.0.1:3032", "127.0.0.1:3033"}
)

type testNode struct {
	selfAddr   string
	host       *NodeHost
	otherPeers []string
	peerGroup  map[string]*Peer
}

// number should be 0-3
func newTestNode(number int) *testNode {
	selfAddr := testAddress[number]
	otherPeers := make([]string, 0)
	for i := 0; i < 4; i++ {
		if i == number {
			continue
		}
		otherPeers = append(otherPeers, testAddress[i])
	}
	name := "TestNode-" + strconv.Itoa(number)
	host := NewNodeHost(name, selfAddr)

	return &testNode{
		selfAddr:   selfAddr,
		host:       host,
		otherPeers: otherPeers,
		peerGroup:  make(map[string]*Peer),
	}
}

func TestNodeHost_ConnectingAndHandleMessage(t *testing.T) {
	var messageCount int
	handleMessage_Print := func(msg Msg) error {
		switch msg.Code {
		case TestMsg:
			var tmsg TestMessage
			rlp.Decode(msg.Payload, &tmsg)
			t.Logf("Message got, content: %s", tmsg.Content)
			messageCount += 1
		default:
			t.Fatal("unknown code")
		}
		return nil
	}

	tnodes := []*testNode{}
	for i := 0; i < 4; i++ {
		tnode := newTestNode(i)
		tnodes = append(tnodes, tnode)
		tnode.host.Start()
	}

	time.Sleep(1 * time.Second) // wait host starts

	for _, tnode := range tnodes {
		tasks := SimpleCreateDialTasks(tnode.selfAddr, tnode.otherPeers, 1, false)
		tnode.host.Dial(tasks)
	}

	time.Sleep(1 * time.Second) // wait dialtask finish

	for _, tnode := range tnodes {
		for _, remoteAddr := range tnode.otherPeers {
			peer, err := tnode.host.BindPeer(remoteAddr, handleMessage_Print)
			if err != nil {
				t.Fatal(err)
			}
			tnode.peerGroup[remoteAddr] = peer
		}
	}

	for _, tnode := range tnodes {
		for remoteAddr, peer := range tnode.peerGroup {
			content := fmt.Sprintf("This message is sent from %s to %s", tnode.selfAddr, remoteAddr)
			tmsg := NewTestMessage(content)
			if err := peer.Send(TestMsg, tmsg); err != nil {
				t.Fatal(err)
			}
			if err := peer.Send(TestMsg, tmsg); err != nil {
				t.Fatal(err)
			}
			if err := peer.Send(TestMsg, tmsg); err != nil {
				t.Fatal(err)
			}
		}
	}

	time.Sleep(2 * time.Second) // wait all messages are received

	if messageCount != 3*12 {
		t.Fatal("Not all messages are got")
	}
}

// Test whether the connection can be lived even long time have not sent messages
func TestNodeHost_LongLiveConnect(t *testing.T) {
	var messageCount int
	handleMessage_Print := func(msg Msg) error {
		switch msg.Code {
		case TestMsg:
			var tmsg TestMessage
			rlp.Decode(msg.Payload, &tmsg)
			t.Logf("Message got, content: %s", tmsg.Content)
			messageCount += 1
		default:
			t.Fatal("unknown code")
		}
		return nil
	}

	tnodes := []*testNode{}
	for i := 0; i < 4; i++ {
		tnode := newTestNode(i)
		tnodes = append(tnodes, tnode)
		tnode.host.Start()
	}

	time.Sleep(1 * time.Second) // wait host starts

	for _, tnode := range tnodes {
		tasks := SimpleCreateDialTasks(tnode.selfAddr, tnode.otherPeers, 1, false)
		tnode.host.Dial(tasks)
	}

	time.Sleep(1 * time.Second) // wait dialtask finish

	for _, tnode := range tnodes {
		for _, remoteAddr := range tnode.otherPeers {
			peer, err := tnode.host.BindPeer(remoteAddr, handleMessage_Print)
			if err != nil {
				t.Fatal(err)
			}
			tnode.peerGroup[remoteAddr] = peer
		}
	}

	t.Log("Now begin sleeping for ten minutes...")

	time.Sleep(10 * time.Minute) // wait for 10 minutes before sending messages

	for _, tnode := range tnodes {
		for remoteAddr, peer := range tnode.peerGroup {
			content := fmt.Sprintf("This message is sent from %s to %s", tnode.selfAddr, remoteAddr)
			tmsg := NewTestMessage(content)
			if err := peer.Send(TestMsg, tmsg); err != nil {
				t.Fatal(err)
			}
		}
	}

	time.Sleep(2 * time.Second) // wait all messages are received

	if messageCount != 12 {
		t.Fatal("Not all messages are got")
	}
}
