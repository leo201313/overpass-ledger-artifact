package network

import (
	"bytes"
	"net"
	"opl/common"
	"testing"
	"time"
)

type testServer struct {
	selfAddr string
	key      *common.KeyPair
}

func TestHandShake_ShouldEnc(t *testing.T) {
	serv1_key, _ := common.GenerateKeyRandom()
	serv2_key, _ := common.GenerateKeyRandom()

	serv1 := &testServer{
		selfAddr: "127.0.0.1:20130",
		key:      serv1_key,
	}

	serv2 := &testServer{
		selfAddr: "127.0.0.1:20131",
		key:      serv2_key,
	}

	go func() {
		time.Sleep(3 * time.Second)
		t.Log("start connecting")
		conn, err := net.Dial("tcp", serv1.selfAddr)
		if err != nil {
			t.Fatal(err)
		}

		srlpxInit := newSRLPX(conn, true, serv2.selfAddr, serv2.key)
		remoteAddr, err := srlpxInit.handshake(serv1.selfAddr)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("enc handshake finished at the initiator side")

		payload1 := []byte(remoteAddr)
		msg1 := Msg{
			Code:       1,
			Size:       uint32(len(payload1)),
			Payload:    bytes.NewReader(payload1),
			ReceivedAt: time.Time{},
		}
		err = srlpxInit.WriteMsg(msg1)
		if err != nil {
			t.Fatal(err)
		}

		msg2, err := srlpxInit.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(msg2.String())

		payload3 := []byte("I've got your message, this is the final check!")
		msg3 := Msg{
			Code:       3,
			Size:       uint32(len(payload3)),
			Payload:    bytes.NewReader(payload3),
			ReceivedAt: time.Time{},
		}

		err = srlpxInit.WriteMsg(msg3)
		if err != nil {
			t.Fatal(err)
		}

	}()

	listen, err := net.Listen("tcp", serv1.selfAddr)
	if err != nil {
		t.Fatal(err)
	}

	finishFlag := make(chan struct{})

	t.Log("start listening")
	conn, err := listen.Accept()
	t.Log("got connected")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		srlpxRecv := newSRLPX(conn, false, serv1.selfAddr, serv1.key)
		remoteAddr, err := srlpxRecv.handshake("")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("enc handshake finished at the receiver side")

		msg1, err := srlpxRecv.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}

		t.Log(msg1.String())

		payload2 := []byte(remoteAddr)
		msg2 := Msg{
			Code:       0,
			Size:       uint32(len(payload2)),
			Payload:    bytes.NewReader(payload2),
			ReceivedAt: time.Time{},
		}
		err = srlpxRecv.WriteMsg(msg2)
		if err != nil {
			t.Fatal(err)
		}

		msg3, err := srlpxRecv.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(msg3.String())

		finishFlag <- struct{}{}

	}()

	<-finishFlag
	t.Log("everything right")
}

func TestHandShake_WithoutEnc(t *testing.T) {
	serv1_key, _ := common.GenerateKeyRandom()
	serv2_key, _ := common.GenerateKeyRandom()

	serv1 := &testServer{
		selfAddr: "127.0.0.1:20130",
		key:      serv1_key,
	}

	serv2 := &testServer{
		selfAddr: "127.0.0.1:20131",
		key:      serv2_key,
	}

	go func() {
		time.Sleep(3 * time.Second)
		t.Log("start connecting")
		conn, err := net.Dial("tcp", serv1.selfAddr)
		if err != nil {
			t.Fatal(err)
		}

		srlpxInit := newSRLPX(conn, false, serv2.selfAddr, serv2.key)
		remoteAddr, err := srlpxInit.handshake(serv1.selfAddr)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("handshake without enc finished at the initiator side")

		payload1 := []byte(remoteAddr)
		msg1 := Msg{
			Code:       2,
			Size:       uint32(len(payload1)),
			Payload:    bytes.NewReader(payload1),
			ReceivedAt: time.Time{},
		}
		err = srlpxInit.WriteMsg(msg1)
		if err != nil {
			t.Fatal(err)
		}

		msg2, err := srlpxInit.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(msg2.String())

		payload3 := []byte("I've got your message, this is the final check!")
		msg3 := Msg{
			Code:       0,
			Size:       uint32(len(payload3)),
			Payload:    bytes.NewReader(payload3),
			ReceivedAt: time.Time{},
		}

		err = srlpxInit.WriteMsg(msg3)
		if err != nil {
			t.Fatal(err)
		}

	}()

	listen, err := net.Listen("tcp", serv1.selfAddr)
	if err != nil {
		t.Fatal(err)
	}

	finishFlag := make(chan struct{})

	t.Log("start listening")
	conn, err := listen.Accept()
	t.Log("got connected")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		srlpxRecv := newSRLPX(conn, false, serv1.selfAddr, serv1.key)
		remoteAddr, err := srlpxRecv.handshake("")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("handshake without enc finished at the receiver side")

		msg1, err := srlpxRecv.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}

		t.Log(msg1.String())

		payload2 := []byte(remoteAddr)
		msg2 := Msg{
			Code:       0,
			Size:       uint32(len(payload2)),
			Payload:    bytes.NewReader(payload2),
			ReceivedAt: time.Time{},
		}
		err = srlpxRecv.WriteMsg(msg2)
		if err != nil {
			t.Fatal(err)
		}

		msg3, err := srlpxRecv.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(msg3.String())

		finishFlag <- struct{}{}

	}()

	<-finishFlag
	t.Log("everything right")
}
