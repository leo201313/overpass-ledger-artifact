package network

import (
	"fmt"
	"github.com/golang/glog"
	"log"
	"opl/coes"
	"opl/logger"
	"opl/rlp"
	"time"
)

// Peer is a wrapped conn that handles for message read and write
type Peer struct {
	ListenAddr     string
	dconn          *dualConn
	msgProcessFunc func(msg Msg) error // this is valid after Peer is bound with abstract instance in the upper protocol

	closed chan struct{}
}

func newPeer(remoteAddr string, dconn *dualConn) *Peer {
	return &Peer{
		ListenAddr: remoteAddr,
		dconn:      dconn,
		closed:     make(chan struct{}),
	}
}

func (p *Peer) Start() {
	go p.readLoop()
	go p.pingLoop()
}

func (p *Peer) pingLoop() {
	ping := time.NewTicker(coes.PingInterval)
	for {
		select {
		case <-ping.C:
			if err := SendItems(p.dconn.sx, pingMsg); err != nil {
				glog.V(logger.Error).Infof("heart Beate wrong with %v", err)
				p.Stop()
				return
			}
		case <-p.closed:
			glog.V(logger.Error).Infoln("pingLoop is closed")
			return
		}
	}
}

func (p *Peer) readLoop() {
	for {
		msg, err := p.dconn.sx.ReadMsg()
		if err != nil {
			glog.V(logger.Error).Infof("fail in readLoop with %v", err)
			p.Stop()
			return
		}
		msg.ReceivedAt = time.Now()
		if err = p.handle(msg); err != nil {
			glog.V(logger.Error).Infof("fail in readLoop with %v", err)
			p.Stop()
			return
		}
	}
}

func (p *Peer) handle(msg Msg) error {
	switch msg.Code {
	case pingMsg:
		msg.Discard()
		go SendItems(p.dconn.sx, pongMsg)
	case discMsg:
		var reason [1]DiscReason
		// This is the last message. We don't need to discard or
		// check errors because, the connection will be closed after it.
		rlp.Decode(msg.Payload, &reason)
		return reason[0]
	case pongMsg:
		return msg.Discard()
	case innerMsg: // TODO: this code is a reserved code for further use
		return msg.Discard()
	default:
		// this means the code is related to upper protocol
		if p.msgProcessFunc != nil {
			return p.msgProcessFunc(msg)
		} else {
			return fmt.Errorf("unknown message with code %d is got and cannot process", msg.Code)
		}
	}
	return nil
}

func (p *Peer) InstallHandleFunc(handleFunc func(msg Msg) error) {
	p.msgProcessFunc = handleFunc
	glog.V(logger.Info).Infof("Peer %s has installed the handle function", p.ListenAddr)
}

func (p *Peer) Stop() {
	close(p.closed)
	//panic("fail in peer connection") // at this version, we use panic to notify error immediately
	log.Printf("connection with peer %s is stopped.", p.ListenAddr)
}

func (p *Peer) Send(msgcode uint64, data interface{}) error {
	return Send(p.dconn.sx, msgcode, data)
}

type PeerGroup struct {
	Peers []*Peer
}

// TODO: the message here is encoded repeatedly and is low efficient.
func (pg *PeerGroup) Broadcast(msgcode uint64, data interface{}) error {
	for i, p := range pg.Peers {
		if err := p.Send(msgcode, data); err != nil {
			return fmt.Errorf("fail in PeerGroup.Broadcast at %dth message which is sent to Peer with address: %s, the details are: %v", i, p.ListenAddr, err)
		}
	}
	return nil
}

func (pg *PeerGroup) AddPeer(p *Peer) {
	pg.Peers = append(pg.Peers, p)
}

func NewPeerGroup() *PeerGroup {
	return &PeerGroup{Peers: make([]*Peer, 0)}
}
