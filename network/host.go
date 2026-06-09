package network

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"opl/coes"
	"opl/common"
	"opl/logger"
	"sync"
	"time"
)

// NodeHost is a server and also a client that should listen requests from
// other nodes and also dial others.
type NodeHost struct {
	Name       string // not important
	ListenAddr string // be like: 127.0.0.1:2013

	listener net.Listener
	dialer   *net.Dialer

	keyPair *common.KeyPair // used for ECDH

	connectedPeers map[string]*Peer // string should be the listen addr of the peer
	peerMux        sync.RWMutex     // to lock peer add and remove

	running bool // whether the host is running, host can only be run only once
	runMux  sync.Mutex

	quit chan struct{}
}

func NewNodeHost(name string, listenAddr string) *NodeHost {
	return &NodeHost{
		Name:           name,
		ListenAddr:     listenAddr,
		keyPair:        nil,
		connectedPeers: nil,
		peerMux:        sync.RWMutex{},
		running:        false,
		runMux:         sync.Mutex{},
		quit:           nil,
	}
}

func (nh *NodeHost) Start() {
	nh.runMux.Lock()
	defer nh.runMux.Unlock()
	if nh.running {
		glog.V(logger.Error).Infoln("NodeHost cannot start due to it has been running once")
		panic("NodeHost cannot start due to it has been running once")
	}

	// initialize the attributes of node host
	keyPair, err := common.GenerateKeyRandom()
	if err != nil {
		glog.V(logger.Error).Infoln(err)
		panic(err)
	}

	nh.keyPair = keyPair
	nh.connectedPeers = make(map[string]*Peer)
	nh.quit = make(chan struct{})

	glog.V(logger.Info).Infof("NodeHost with name: %s, listen address: %s is initialized and ready to start...", nh.Name, nh.ListenAddr)

	nh.startListening()
	glog.V(logger.Info).Infoln("NodeHost start listening")

	nh.running = true
	glog.V(logger.Info).Infoln("NodeHost has been launched")
}

func (nh *NodeHost) startListening() {
	// Launch the TCP listener
	listener, err := net.Listen("tcp", nh.ListenAddr)
	if err != nil {
		glog.V(logger.Error).Infof("node host cannot start listening: %v", err)
		panic(err)
	}
	nh.listener = listener
	go nh.listenLoop()
}

// listenLoop runs in its own goroutine and accepts
// inbound connections.
func (nh *NodeHost) listenLoop() {
	glog.V(logger.Info).Infof("Listening on %s", nh.ListenAddr)
	for {
		fd, err := nh.listener.Accept()
		if err != nil {
			glog.V(logger.Error).Infof("Can't Accept TCP Connection Request,err:%v", err)
			continue
		}
		go nh.setupConn(fd, "") // listener does not know the remote address
	}
}

// addPeer adds a peer into host node and also check whether it has been connected
func (nh *NodeHost) addPeer(peer *Peer) {
	nh.peerMux.Lock()
	defer nh.peerMux.Unlock()
	remoteAddr := peer.ListenAddr
	if _, ok := nh.connectedPeers[remoteAddr]; ok {
		glog.V(logger.Warn).Infof("remote peer %s has already been connected", remoteAddr)
	}
	nh.connectedPeers[remoteAddr] = peer
	glog.V(logger.Info).Infof("remote peer %s has been connected and added to the peer map", remoteAddr)
}

func (nh *NodeHost) setupConn(conn net.Conn, remoteAddr string) error {
	sx := newSRLPX(conn, false, nh.ListenAddr, nh.keyPair)
	remoteAddr, err := sx.handshake(remoteAddr)
	if err != nil {
		glog.V(logger.Error).Infof("Can't connect with node %s, and details are %v", remoteAddr, err)
		return err
	}
	glog.V(logger.Info).Infof("succeed in handshaking with node %s", remoteAddr)
	dconn := dualConn{
		fd:         conn,
		sx:         sx,
		remoteAddr: remoteAddr,
	}

	peer := newPeer(remoteAddr, &dconn)
	peer.Start()

	nh.addPeer(peer)
	glog.V(logger.Info).Infof("connection with address %s has been setup", remoteAddr)
	return nil

}

// at this version, Dial will do the dial task one by one
func (nh *NodeHost) Dial(tasks []DialTask) {
	nh.runMux.Lock()
	defer nh.runMux.Unlock()
	if !nh.running {
		glog.V(logger.Error).Infoln("node hosr dial without running!")
		panic("node hosr dial without running!")
	}

	if nh.dialer == nil {
		nh.dialer = &net.Dialer{Timeout: coes.DefaultDialTimeout}
	}

	// do task one by one
	for _, task := range tasks {
		err := nh.doDialTask(task)
		if err != nil {
			glog.V(logger.Error).Infoln(err)
			panic(err) // at this time, we force the node host to finish their dial tasks completely
		}
	}
}

func (nh *NodeHost) doDialTask(task DialTask) error {
	if len(task.remoteAddr) == 0 {
		return fmt.Errorf("fail in doDialTask: remote address is none")
	}

	// try at least once
	if task.maxTry <= 0 {
		task.maxTry = 1
	}

	var conn net.Conn
	var err error

	for i := 0; i < task.maxTry; i++ {
		conn, err = nh.dialer.Dial("tcp", task.remoteAddr)
		if err != nil {
			glog.V(logger.Warn).Infof("fail in do dial task with %s, details are: %v", task.remoteAddr, err)
			time.Sleep(coes.ReTryDialWait)
			continue
		}

		// do slpx handshake
		err = nh.setupConn(conn, task.remoteAddr)
		if err != nil {
			conn = nil
			continue
		}

		return nil
	}

	glog.V(logger.Warn).Infof("max try runs out: fail in do dial task with %s", task.remoteAddr)
	return fmt.Errorf("max try runs out: fail in do dial task with %s", task.remoteAddr)
}

// Stop terminates the node host and all active peer connections.
// Note: node host can only be launched only once, which means the node host cannot
// be start again after stop
func (nh *NodeHost) Stop() {
	nh.runMux.Lock()
	defer nh.runMux.Unlock()
	if !nh.running {
		return
	}
	nh.running = false
	if nh.listener != nil {
		nh.listener.Close()
	}
	close(nh.quit)
}

// BindPeer check the peers' connection and bind it to abstract instance in the upper protocol
func (nh *NodeHost) BindPeer(remoteAddr string, msgProcessFunc func(msg Msg) error) (*Peer, error) {
	nh.peerMux.RLock()
	defer nh.peerMux.RUnlock()
	if peer, ok := nh.connectedPeers[remoteAddr]; !ok {
		return nil, fmt.Errorf("cannot bind peer with %s as it has not been connected", remoteAddr)
	} else {
		peer.InstallHandleFunc(msgProcessFunc)
		glog.V(logger.Info).Infof("Peer %s has bind successfully", remoteAddr)
		return peer, nil
	}
}
