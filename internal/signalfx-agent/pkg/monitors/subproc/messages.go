package subproc

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

// MessageType is the type of the message going to the subprocess runner
type MessageType uint32

// MessageTypes that are standard
const (
	MessageTypeNone            MessageType = 0
	MessageTypeConfigure       MessageType = 1
	MessageTypeConfigureResult MessageType = 2
	MessageTypeShutdown        MessageType = 3
	MessageTypeLog             MessageType = 4
)

type configResult struct {
	Error *string `json:"error"`
}

type messageReadWriter struct {
	Reader            io.ReadCloser
	Writer            io.WriteCloser
	lastPayloadReader *io.LimitedReader
}

// MessageReceiver can get messages from subprocesses
type MessageReceiver interface {
	RecvMessage() (MessageType, io.Reader, error)
}

// RecvMessage blocks until it receives a complete message from the Reader
// pipe.  This is not thread-safe.
func (m *messageReadWriter) RecvMessage() (MessageType, io.Reader, error) {
	if m.lastPayloadReader != nil && m.lastPayloadReader.N > 0 {
		return MessageTypeNone, nil, errors.New("last payload was not fully read, cannot receive")
	}

	var buf [4]byte

	if _, err := io.ReadFull(m.Reader, buf[:]); err != nil {
		return MessageTypeNone, nil, err
	}

	msgType := MessageType(binary.BigEndian.Uint32(buf[:]))

	if _, err := io.ReadFull(m.Reader, buf[:]); err != nil {
		return MessageTypeNone, nil, err
	}

	size := binary.BigEndian.Uint32(buf[:])

	payloadReader := &io.LimitedReader{
		R: m.Reader,
		N: int64(size),
	}
	m.lastPayloadReader = payloadReader

	return msgType, payloadReader, nil
}

// SendMessage frames and sends the given message to the Writer.  This is not
// thread-safe.
func (m *messageReadWriter) SendMessage(msgType MessageType, payload []byte) error {
	var buf [4]byte

	logrus.Debugf("Sending message to subproc (type %d) with length %d", msgType, len(payload))

	binary.BigEndian.PutUint32(buf[:], uint32(msgType))
	if _, err := m.Writer.Write(buf[:]); err != nil {
		return err
	}

	binary.BigEndian.PutUint32(buf[:], uint32(len(payload)))
	if _, err := m.Writer.Write(buf[:]); err != nil {
		return err
	}

	if payload != nil {
		_, err := m.Writer.Write(payload)
		return err
	}
	return nil
}

func (m *messageReadWriter) Close() {
	m.Reader.Close()
	m.Writer.Close()
}
