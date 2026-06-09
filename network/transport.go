package network

import "net"

type transport interface {
	// The handshake is used to negotiate the encryption secrets (or not, maybe)
	handshake(DestAddr string) (Src string, err error)

	// The MsgReadWriter can only be used after the encryption handshake has completed.
	MsgReadWriter

	close(err error)
}

// a wrapper of net.Conn and srlpx, used for Peer
type dualConn struct {
	fd         net.Conn
	sx         *srlpx
	remoteAddr string // valid after handshake
}
