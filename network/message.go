package network

import (
	"fmt"
	"io"
	"opl/rlp"
	"time"
)

// MsgReadWriter provides reading and writing of encoded messages.
// Implementations should ensure that ReadMsg and WriteMsg can be
// called simultaneously from multiple goroutines.
type MsgReadWriter interface {
	MsgReader
	MsgWriter
}

type MsgReader interface {
	ReadMsg() (Msg, error)
}

type MsgWriter interface {
	// WriteMsg sends a message. It will block until the message's
	// Payload has been consumed by the other end.
	//
	// Note that messages can be sent only once because their
	// payload reader is drained.
	WriteMsg(Msg) error
}

// Send writes an RLP-encoded message with the given code.
// data should encode as an RLP list.
func Send(w MsgWriter, msgcode uint64, data interface{}) error {
	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}
	return w.WriteMsg(Msg{Code: msgcode, Size: uint32(size), Payload: r})
}

// SendItems writes an RLP with the given code and data elements.
// For a call such as:
//
//	SendItems(w, code, e1, e2, e3)
//
// the message payload will be an RLP list containing the items:
//
//	[e1, e2, e3]
func SendItems(w MsgWriter, msgcode uint64, elems ...interface{}) error {
	return Send(w, msgcode, elems)
}

// Msg defines the structure of a basic network message.
//
// Note that a Msg can only be sent once since the Payload reader is
// consumed during sending. It is not possible to create a Msg and
// send it any number of times. If you want to reuse an encoded
// structure, encode the payload into a byte array and create a
// separate Msg with a bytes.Reader as Payload for each send.
type Msg struct {
	Code       uint64
	Size       uint32
	Payload    io.Reader
	ReceivedAt time.Time
}

const MsgCodeLen = 8 // byte len

func (m *Msg) String() string {
	str := fmt.Sprintf("Msg.Code: %d\n", m.Code)
	str += fmt.Sprintf("Msg.Size: %d\n", m.Size)
	buff := make([]byte, m.Size)
	m.Payload.Read(buff)
	str += fmt.Sprintf("Msg.Payload: %s", string(buff))
	return str
}

// Discard reads any remaining payload data into a black hole.
func (msg Msg) Discard() error {
	_, err := io.Copy(io.Discard, msg.Payload)
	return err
}

// Decode parses the RLP content of a message into
// the given value, which must be a pointer.
//
// For the decoding rules, please see package rlp.
func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		return err
	}
	return nil
}

type TestMessage struct {
	Content string
}

func NewTestMessage(content string) *TestMessage {
	return &TestMessage{Content: content}
}
