package protocol

import (
	"errors"
	"io"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/util"
)

var (
	ErrorInvalidMessage  = errors.New("invalid message")
	ErrorInvalidHeader   = errors.New("invalid header")
	ErrorReadOutofRange  = errors.New("read body length out of range")
	ErrorWriteOutofRange = errors.New("write body length out of range")

	bufPool = util.NewBufferPool()
)

// 有一个完整的消息包含头和body两部分
type Message struct {
	Header ProtocolHeader
	Body   ProtocolBody
}

type DataCreator interface {
	CreateHeader() ProtocolHeader
	CreateBody(cmd int) ProtocolBody
}

func ReadMessage(reader io.Reader, dc DataCreator) (*Message, error) {
	return ReadLimitMessage(reader, dc, 0)
}

// 给定消息结构读取
func ReadLimitMessage(reader io.Reader, dc DataCreator, limitSize int) (*Message, error) {
	header := dc.CreateHeader()

	headerLength := header.Length()
	buff := make([]byte, headerLength)
	_, err := io.ReadFull(reader, buff)
	if err != nil {
		return nil, err
	}

	err = header.Decode(buff)
	if err != nil {
		return nil, err
	}

	bodyLength := header.BodyLength()
	if bodyLength < 0 || (limitSize > 0 && bodyLength > limitSize) {
		log.Warnf("invalid header length: %d", bodyLength)
		return nil, ErrorReadOutofRange
	}

	body := dc.CreateBody(header.Cmd())
	if bodyLength > 0 {
		buff = make([]byte, bodyLength)
		_, err = io.ReadFull(reader, buff)
		if err != nil {
			return nil, err
		}
		err = body.Decode(buff)
	}
	return &Message{Header: header, Body: body}, err
}

// 编码Message
func UnmarshalMessgae(b []byte, dc DataCreator) (*Message, error) {
	message := &Message{
		Header: dc.CreateHeader(),
	}

	headerLength := message.Header.Length()

	if len(b) < headerLength {
		return message, ErrorReadOutofRange
	}
	head := b[:headerLength]
	err := message.Header.Decode(head)
	if err != nil {
		return message, err
	}

	if message.Header.BodyLength() != len(b)-headerLength {
		return message, ErrorInvalidMessage
	}
	message.Body = dc.CreateBody(message.Header.Cmd())
	err = message.Body.Decode(b[headerLength:])
	return message, err
}

// write 由服务端自己控制,不用限制字数
func WriteMessage(conn io.Writer, message *Message) error {
	return WriteLimitMessage(conn, message, 0) // unlimited
}

func WriteLimitMessage(conn io.Writer, message *Message, limitSize int) error {
	if message.Header == nil {
		return ErrorInvalidMessage
	}
	data, err := MarshalLimitMessgae(message, limitSize)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

func MarshalLimitMessgae(message *Message, limitSize int) ([]byte, error) {
	if message.Header == nil {
		return nil, ErrorInvalidHeader
	}

	var body []byte
	var err error
	if message.Body != nil {
		body, err = message.Body.Encode()
		if err != nil {
			return nil, err
		}
	}
	if limitSize > 0 && len(body) > limitSize {
		return nil, ErrorWriteOutofRange
	}
	message.Header.SetBodyLength(len(body))
	buffer := bufPool.Get()
	defer bufPool.Put(buffer)

	hdr, err := message.Header.Encode()
	if err != nil {
		return nil, err
	}

	buffer.Write(hdr)
	buffer.Write(body)
	return buffer.Bytes(), nil
}

// 编码Message
func MarshalMessgae(message *Message) ([]byte, error) {
	return MarshalLimitMessgae(message, 0)
}
