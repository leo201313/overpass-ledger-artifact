package network

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"hash"
	"io"
	"net"
	"opl/coes"
	"opl/common"
	"opl/utils"
	"sync"
	"time"
)

const (
	maxUint24 = ^uint32(0) >> 8
)

const (
	authLen = coes.PUBKEY_LENGTH*2 + coes.SIG_LENGTH
)

// srlpx is the transport protocol used by actual (non-test) connections.
// It wraps the frame encoder with locks and read/write deadlines.
// srlpx is short for simple rlp transport.
type srlpx struct {
	shouldEnc bool

	selfAddr string
	key      *common.KeyPair

	fd net.Conn

	rmu, wmu sync.Mutex
	rw       MsgReadWriter
}

func newSRLPX(fd net.Conn, shouldEnc bool, selfAddr string, key *common.KeyPair) *srlpx {
	return &srlpx{
		shouldEnc: shouldEnc,
		selfAddr:  selfAddr,
		key:       key,
		fd:        fd,
		rmu:       sync.Mutex{},
		wmu:       sync.Mutex{},
		rw:        nil,
	}
}

func (s *srlpx) ReadMsg() (Msg, error) {
	s.rmu.Lock()
	defer s.rmu.Unlock()

	s.fd.SetReadDeadline(time.Now().Add(coes.FrameReadTimeout))
	return s.rw.ReadMsg()
}

func (s *srlpx) WriteMsg(msg Msg) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	s.fd.SetWriteDeadline(time.Now().Add(coes.FrameWriteTimeout))
	return s.rw.WriteMsg(msg)
}

func (s *srlpx) close(err error) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	// Tell the remote end why we're disconnecting if possible.
	if s.rw != nil {
		if r, ok := err.(DiscReason); ok && r != DiscNetworkError {
			s.fd.SetWriteDeadline(time.Now().Add(coes.DiscWriteTimeout))
			SendItems(s.rw, discMsg, r)
		}
	}
	s.fd.Close()
}

// secrets represents the connection secrets
// which are negotiated during the encryption handshake.
type secrets struct {
	ShouldEnc             bool   // tells whether encryption is in need
	RemoteADDR            string // here it is the ip addr of the remote peer
	AES, MAC              []byte
	EgressMAC, IngressMAC hash.Hash
}

var (
	// this is used in place of actual frame header data.
	zeroHeader = []byte{0xC2, 0x80, 0x80}
	// sixteen zero bytes
	zero16 = make([]byte, 16)

	// this is used to fill out the headerbuff of srlpxFrameRW_WithoutEnc, it is 5 bytes
	cheackByets = []byte{0xC2, 0x80, 0x80, 0x80, 0x81}

	// this is used to fill out the headerbuff of srlpxFrameRW_Big, it is 5 bytes
	cheackByets4 = []byte{0xC2, 0x80, 0x80, 0x81}
)

// encHandshake contains the state of the encryption handshake.
type encHandshake struct {
	initiator  bool
	remoteAddr string
	shouldEnc  bool

	remotePub              [coes.PUBKEY_LENGTH]byte // remote-pubk
	selfNonce, remoteNonce [coes.PUBKEY_LENGTH]byte // nonce
}

func (enchs *encHandshake) secrets(key *common.KeyPair) (s secrets, err error) {
	sharedNonce, err := utils.XorBytes(enchs.remoteNonce[:], enchs.selfNonce[:])
	if err != nil {
		return s, err
	}

	s.ShouldEnc = enchs.shouldEnc
	s.RemoteADDR = enchs.remoteAddr

	aes, err := key.ComputeSharedAES(enchs.remotePub, sharedNonce)
	if err != nil {
		return s, err
	}

	mac := sha256.Sum256(aes[:])
	s.AES = aes[:]
	s.MAC = mac[:]

	mac1 := sha256.New()
	mac1.Write(append(bytes.Clone(mac[:]), sharedNonce...))

	mac2 := sha256.New()
	mac2.Write(append(bytes.Clone(mac[:]), sharedNonce...))

	s.EgressMAC, s.IngressMAC = mac1, mac2

	return s, nil
}

// HeadMessage is used to tell the remote the bind IP addr and whether it should use encrypted connection
type HeadMessage struct {
	IPAddr    string // be like: 127.0.0.1:2013
	ShouldEnc bool
}

// handshake runs the handshake to negotiate a shared secret
// for following encrypted message transferring.
func (s *srlpx) handshake(DestAddr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), coes.HandshakeTimeout)
	defer cancel()

	var (
		sec  secrets
		err  error
		done = make(chan struct{})
	)

	go func() {
		defer close(done)
		if len(DestAddr) == 0 { // wait to response the remote peer
			sec, err = s.receiverHandshake()
		} else { // initiate a handshake
			sec, err = s.initiatorHandshake(DestAddr)
		}
	}()

	select {
	case <-ctx.Done():
		return DestAddr, ctx.Err() // Return error if timeout
	case <-done:
		// Proceed if handshake is completed within 2 seconds
		if err != nil {
			return DestAddr, err
		}
	}

	s.wmu.Lock()
	if sec.ShouldEnc {
		s.rw = newSRLPXFrameRW(s.fd, sec)
	} else {
		// s.rw = newSrlpxFrameRW_WithoutEnc(s.fd)
		s.rw = newSrlpxFrameRW_Big(s.fd)
	}
	s.wmu.Unlock()
	return sec.RemoteADDR, nil
}

// receiverHandshake negotiates a session token on conn.
// it should be called on the listening side of the connection.
func (sx *srlpx) receiverHandshake() (s secrets, err error) {
	var addrBuf [256]byte
	n, err := sx.fd.Read(addrBuf[:])
	if err != nil {
		return s, err
	}

	var remoteHead HeadMessage
	dec := gob.NewDecoder(bytes.NewReader(addrBuf[:n]))
	dec.Decode(&remoteHead)

	s.RemoteADDR = remoteHead.IPAddr
	sx.shouldEnc = remoteHead.ShouldEnc // the receiver doesn't know whether it needs encrypted connection

	// exchange the head message
	var selfHead HeadMessage
	selfHead.ShouldEnc = sx.shouldEnc // it follows the remote peer
	selfHead.IPAddr = sx.selfAddr

	var tempBuffer bytes.Buffer
	enc := gob.NewEncoder(&tempBuffer)
	enc.Encode(&selfHead)

	_, err = sx.fd.Write(tempBuffer.Bytes())
	if err != nil {
		return s, err
	}

	// read remote auth message
	remoteAuth := make([]byte, authLen)
	if _, err := io.ReadFull(sx.fd, remoteAuth); err != nil { //读取授权消息
		return s, err
	}

	remoteNonce, remotePubkey, err := decodeAuthMessage(remoteAuth)
	if err != nil {
		return s, err
	}

	if !sx.verifyRemote(remoteHead.IPAddr, remotePubkey, remoteHead.ShouldEnc) {
		return s, fmt.Errorf("fail in srlpx.receiverEncHandshake: unmatch remote IP addr and public key")
	}

	// exchange nonce
	var selfNonce [coes.PUBKEY_LENGTH]byte
	if _, err := rand.Read(selfNonce[:]); err != nil {
		return s, err
	}

	selfAuth := sx.buildAuthMessage(selfNonce)
	_, err = sx.fd.Write(selfAuth)
	if err != nil {
		return s, err
	}

	enchs := encHandshake{
		initiator:   false,
		remoteAddr:  remoteHead.IPAddr,
		shouldEnc:   sx.shouldEnc,
		remotePub:   remotePubkey,
		selfNonce:   selfNonce,
		remoteNonce: remoteNonce,
	}

	s, err = enchs.secrets(sx.key)
	if err != nil {
		return s, err
	}

	return s, nil
}

// TODO: verifyRemote verifies the IP addr and the public key of remote peer and also whether
// should use encrypted connection. In real situations, the public key (or say the certifications) are allocated to different nodes
// based on pre-defined configures through off-line negotiations. However, at this time it
// only check whether the encrypted connection is in need.
func (s *srlpx) verifyRemote(remoteAddr string, pubkey [coes.PUBKEY_LENGTH]byte, shouldEnc bool) bool {
	return s.shouldEnc == shouldEnc
}

// initiatorEncHandshake initiates a handshake.
// it should be called on the dialing side of the connection.
func (sx *srlpx) initiatorHandshake(remoteAddr string) (s secrets, err error) {
	enchs, err := sx.newInitiatorHandshake(remoteAddr)
	if err != nil {
		return s, err
	}
	secret, err := enchs.secrets(sx.key)
	return secret, err
}

// buildAuthMessage builds an auth message to exchange nonce with the remote peer.
// The combination be like: pubkey + nonce + signature of nonce
func (s *srlpx) buildAuthMessage(nonce [coes.PUBKEY_LENGTH]byte) []byte {
	signature, _ := s.key.Sign(nonce[:])

	auth := make([]byte, authLen)

	copy(auth[:coes.PUBKEY_LENGTH], s.key.PubKeyBytes[:])
	copy(auth[coes.PUBKEY_LENGTH:2*coes.PUBKEY_LENGTH], nonce[:])
	copy(auth[2*coes.PUBKEY_LENGTH:], signature[:])

	return auth
}

// decodeAuthMessage decodes the auth message
func decodeAuthMessage(auth []byte) (nonce, pubkey [coes.PUBKEY_LENGTH]byte, err error) {
	signature := [coes.SIG_LENGTH]byte{}
	copy(pubkey[:], auth[:coes.PUBKEY_LENGTH])
	copy(nonce[:], auth[coes.PUBKEY_LENGTH:2*coes.PUBKEY_LENGTH])
	copy(signature[:], auth[2*coes.PUBKEY_LENGTH:])
	if common.VerifySignature(nonce[:], pubkey, signature) {
		return nonce, pubkey, nil
	} else {
		return nonce, pubkey, fmt.Errorf("invalid auth message is got")
	}
}

func (s *srlpx) newInitiatorHandshake(remoteAddr string) (*encHandshake, error) {
	// send the head message
	var selfHead HeadMessage
	selfHead.ShouldEnc = s.shouldEnc // the initiator determines whether this connection should be encrypted
	selfHead.IPAddr = s.selfAddr

	var tempBuffer bytes.Buffer
	enc := gob.NewEncoder(&tempBuffer)
	enc.Encode(&selfHead)

	_, err := s.fd.Write(tempBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// read the remote head message
	var addrBuf [256]byte
	n, err := s.fd.Read(addrBuf[:])
	if err != nil {
		return nil, err
	}

	var remoteHead HeadMessage
	dec := gob.NewDecoder(bytes.NewReader(addrBuf[:n]))
	dec.Decode(&remoteHead)

	if remoteHead.IPAddr != remoteAddr {
		return nil, fmt.Errorf("fail in srlpx.newInitiatorHandshake: the remoteAddr should be %s, but remote peer back is: %s", remoteAddr, remoteHead.IPAddr)
	}

	// build self auth message
	var selfNonce [coes.PUBKEY_LENGTH]byte
	if _, err := rand.Read(selfNonce[:]); err != nil {
		return nil, err
	}

	selfAuth := s.buildAuthMessage(selfNonce)
	_, err = s.fd.Write(selfAuth)
	if err != nil {
		return nil, err
	}

	//// NOTE: this wait is very important.
	//time.Sleep(200 * time.Millisecond)

	// read remote auth message
	remoteAuth := make([]byte, authLen)
	if _, err := io.ReadFull(s.fd, remoteAuth); err != nil { //读取授权消息
		return nil, err
	}

	remoteNonce, remotePubkey, err := decodeAuthMessage(remoteAuth)
	if err != nil {
		return nil, err
	}

	if !s.verifyRemote(remoteAddr, remotePubkey, s.shouldEnc) {
		return nil, fmt.Errorf("fail in srlpx.newInitiatorHandshake: unmatch remote IP addr and public key")
	}

	enchs := encHandshake{
		initiator:   true,
		remoteAddr:  remoteAddr,
		shouldEnc:   s.shouldEnc,
		remotePub:   remotePubkey,
		selfNonce:   selfNonce,
		remoteNonce: remoteNonce,
	}

	return &enchs, nil

}

// srlpxFrameRW implements a simplified version of rlpx framing.
// chunked messages are not supported and all headers are equal to
// zeroHeader.
// srlpxFrameRW is not safe for concurrent use from multiple goroutines.
type srlpxFrameRW struct {
	conn io.ReadWriter

	enc cipher.Stream
	dec cipher.Stream

	macCipher  cipher.Block
	egressMAC  hash.Hash
	ingressMAC hash.Hash
}

// create a new srlpxFrameRW with encryption
func newSRLPXFrameRW(conn io.ReadWriter, s secrets) *srlpxFrameRW {
	macc, err := aes.NewCipher(s.MAC)
	if err != nil {
		panic("invalid MAC secret: " + err.Error())
	}
	encc, err := aes.NewCipher(s.AES)
	if err != nil {
		panic("invalid AES secret: " + err.Error())
	}
	// we use an all-zeroes IV for AES because the key used
	// for encryption is ephemeral.
	iv := make([]byte, encc.BlockSize())
	return &srlpxFrameRW{
		conn:       conn,
		enc:        cipher.NewCTR(encc, iv),
		dec:        cipher.NewCTR(encc, iv),
		macCipher:  macc,
		egressMAC:  s.EgressMAC,
		ingressMAC: s.IngressMAC,
	}
}

// updateMAC reseeds the given hash with encrypted seed.
// it returns the first 16 bytes of the hash sum after seeding.
func updateMAC(mac hash.Hash, block cipher.Block, seed []byte) []byte {
	aesbuf := make([]byte, aes.BlockSize)
	block.Encrypt(aesbuf, mac.Sum(nil))
	for i := range aesbuf {
		aesbuf[i] ^= seed[i]
	}
	mac.Write(aesbuf)
	return mac.Sum(nil)[:16]
}

func readInt24(b []byte) uint32 {
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

func readInt32(b []byte) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func putInt24(v uint32, b []byte) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func putInt32(v uint32, b []byte) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func (sw *srlpxFrameRW) WriteMsg(msg Msg) error {
	// write header
	headbuf := make([]byte, 32)
	fsize := uint32(MsgCodeLen) + msg.Size
	if fsize > maxUint24 {
		return errors.New("message size overflows uint24")
	}
	putInt24(fsize, headbuf) // TODO: check overflow
	copy(headbuf[3:], zeroHeader)
	sw.enc.XORKeyStream(headbuf[:16], headbuf[:16]) // first half is now encrypted

	// write header MAC
	copy(headbuf[16:], updateMAC(sw.egressMAC, sw.macCipher, headbuf[:16]))
	if _, err := sw.conn.Write(headbuf); err != nil {
		return err
	}

	// write encrypted frame, updating the egress MAC hash with
	// the data written to conn.
	tee := cipher.StreamWriter{S: sw.enc, W: io.MultiWriter(sw.conn, sw.egressMAC)}
	ptype := utils.Uint64ToBytes(msg.Code)
	if _, err := tee.Write(ptype); err != nil {
		return err
	}
	if _, err := io.Copy(tee, msg.Payload); err != nil {
		return err
	}
	if padding := fsize % 16; padding > 0 {
		if _, err := tee.Write(zero16[:16-padding]); err != nil {
			return err
		}
	}

	// write frame MAC. egress MAC hash is up to date because
	// frame content was written to it as well.
	fmacseed := sw.egressMAC.Sum(nil)
	mac := updateMAC(sw.egressMAC, sw.macCipher, fmacseed)
	_, err := sw.conn.Write(mac)
	return err
}

func (sw *srlpxFrameRW) ReadMsg() (msg Msg, err error) {
	// read the header
	headbuf := make([]byte, 32)
	if _, err := io.ReadFull(sw.conn, headbuf); err != nil {
		return msg, err
	}
	// verify header mac
	shouldMAC := updateMAC(sw.ingressMAC, sw.macCipher, headbuf[:16])
	if !hmac.Equal(shouldMAC, headbuf[16:]) {
		return msg, errors.New("bad header MAC")
	}
	sw.dec.XORKeyStream(headbuf[:16], headbuf[:16]) // first half is now decrypted
	fsize := readInt24(headbuf)
	// ignore protocol type for now

	// read the frame content
	var rsize = fsize // frame size rounded up to 16 byte boundary
	if padding := fsize % 16; padding > 0 {
		rsize += 16 - padding
	}
	framebuf := make([]byte, rsize)
	if _, err := io.ReadFull(sw.conn, framebuf); err != nil {
		return msg, err
	}

	// read and validate frame MAC. we can re-use headbuf for that.
	sw.ingressMAC.Write(framebuf)
	fmacseed := sw.ingressMAC.Sum(nil)
	if _, err := io.ReadFull(sw.conn, headbuf[:16]); err != nil {
		return msg, err
	}
	shouldMAC = updateMAC(sw.ingressMAC, sw.macCipher, fmacseed)
	if !hmac.Equal(shouldMAC, headbuf[:16]) {
		return msg, errors.New("bad frame MAC")
	}

	// decrypt frame content
	sw.dec.XORKeyStream(framebuf, framebuf)

	// decode message code

	ptype := framebuf[:MsgCodeLen]
	payload := framebuf[MsgCodeLen:fsize]

	code, err := utils.BytesToUint64(ptype)
	if err != nil {
		return msg, err
	}
	msg.Code = code
	msg.Size = uint32(len(payload))
	msg.Payload = bytes.NewReader(payload)
	return msg, nil
}

// srlpxFrameRW_WithoutEnc implements a simplified version of rlpx framing.
// chunked messages are not supported and all headers are equal to
// zeroHeader. srlpxFrameRW_WithoutEnc use no crypto technology to transfer data.
// srlpxFrameRW_WithoutEnc is not safe for concurrent use from multiple goroutines.
type srlpxFrameRW_WithoutEnc struct {
	conn io.ReadWriter
}

// create a new srlpxFrameRW_WithoutEnc
func newSrlpxFrameRW_WithoutEnc(conn io.ReadWriter) *srlpxFrameRW_WithoutEnc {
	return &srlpxFrameRW_WithoutEnc{conn: conn}
}

func (sw *srlpxFrameRW_WithoutEnc) WriteMsg(msg Msg) error {
	ptype := utils.Uint64ToBytes(msg.Code) // must be 8 bytes

	// write header
	headbuf := make([]byte, 16)
	fsize := msg.Size
	if fsize > maxUint24 {
		return errors.New("message size overflows uint24")
	}
	putInt24(msg.Size, headbuf)     // TODO: check overflow
	copy(headbuf[3:11], ptype)      // encode the msg code into headbuf
	copy(headbuf[11:], cheackByets) // fill in the headbuf

	if _, err := sw.conn.Write(headbuf); err != nil {
		return err
	}

	if _, err := io.Copy(sw.conn, msg.Payload); err != nil {
		return err
	}

	if padding := fsize % 16; padding > 0 {
		if _, err := sw.conn.Write(zero16[:16-padding]); err != nil {
			return err
		}
	}

	return nil
}

func (sw *srlpxFrameRW_WithoutEnc) ReadMsg() (msg Msg, err error) {
	// read the header
	headbuf := make([]byte, 16)
	if _, err := io.ReadFull(sw.conn, headbuf); err != nil {
		return msg, err
	}
	fsize := readInt24(headbuf)

	msgCode, err := utils.BytesToUint64(headbuf[3:11])
	if err != nil {
		return msg, err
	}

	if !bytes.Equal(headbuf[11:], cheackByets) {
		return msg, errors.New("cannot decode message as the srlpxFrameRW_WithoutEnc headbuf")
	}

	// read the frame content
	var rsize = fsize // frame size rounded up to 16 byte boundary
	if padding := fsize % 16; padding > 0 {
		rsize += 16 - padding
	}
	framebuf := make([]byte, rsize)
	if _, err := io.ReadFull(sw.conn, framebuf); err != nil {
		return msg, err
	}

	// decode message code
	content := bytes.NewReader(framebuf[:fsize])

	msg.Code = msgCode
	msg.Size = uint32(content.Len())
	msg.Payload = content
	return msg, nil
}

// srlpxFrameRW_Big implements a simplified version of rlpx framing
// and supports message size that over uint24 to uint32.
// chunked messages are not supported and all headers are equal to
// zeroHeader. srlpxFrameRW_Big use no crypto technology to transfer data.
// srlpxFrameRW_Big is not safe for concurrent use from multiple goroutines.
type srlpxFrameRW_Big struct {
	conn io.ReadWriter
}

// create a new srlpxFrameRW_Big
func newSrlpxFrameRW_Big(conn io.ReadWriter) *srlpxFrameRW_Big {
	return &srlpxFrameRW_Big{conn: conn}
}

func (sb *srlpxFrameRW_Big) WriteMsg(msg Msg) error {
	ptype := utils.Uint64ToBytes(msg.Code) // must be 8 bytes

	// write header
	headbuf := make([]byte, 16)
	fsize := msg.Size
	putInt32(msg.Size, headbuf)      // TODO: check overflow
	copy(headbuf[4:12], ptype)       // encode the msg code into headbuf
	copy(headbuf[12:], cheackByets4) // fill in the headbuf

	if _, err := sb.conn.Write(headbuf); err != nil {
		return err
	}

	if _, err := io.Copy(sb.conn, msg.Payload); err != nil {
		return err
	}

	if padding := fsize % 16; padding > 0 {
		if _, err := sb.conn.Write(zero16[:16-padding]); err != nil {
			return err
		}
	}

	return nil
}

func (sb *srlpxFrameRW_Big) ReadMsg() (msg Msg, err error) {
	// read the header
	headbuf := make([]byte, 16)
	if _, err := io.ReadFull(sb.conn, headbuf); err != nil {
		return msg, err
	}
	fsize := readInt32(headbuf)

	msgCode, err := utils.BytesToUint64(headbuf[4:12])
	if err != nil {
		return msg, err
	}

	if !bytes.Equal(headbuf[12:], cheackByets4) {
		return msg, errors.New("cannot decode message as the srlpxFrameRW_Big headbuf")
	}

	// read the frame content
	var rsize = fsize // frame size rounded up to 16 byte boundary
	if padding := fsize % 16; padding > 0 {
		rsize += 16 - padding
	}
	framebuf := make([]byte, rsize)
	if _, err := io.ReadFull(sb.conn, framebuf); err != nil {
		return msg, err
	}

	// decode message code
	content := bytes.NewReader(framebuf[:fsize])

	msg.Code = msgCode
	msg.Size = uint32(content.Len())
	msg.Payload = content
	return msg, nil
}
