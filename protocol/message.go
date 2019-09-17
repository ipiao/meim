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

// 协议头
type ProrocolHeader interface {
	ProtocolData
	Cmd() int // 协议指令
	BodyLength() int
	SetBodyLength(n int)
}

// 有一个完整的消息包含头和body两部分
type Message struct {
	Header ProrocolHeader
	Body   ProtocolBody
}

type MessageCodec interface {
	HeaderLength() int
	DecodeHeader(b []byte) (ProrocolHeader, error)
	DecodeBody(cmd int, body []byte) (ProtocolBody, error)
}

// 给定消息结构读取
func ReadLimitMessage(reader io.Reader, codec MessageCodec, limitSize int) (*Message, error) {
	headLength := codec.HeaderLength()
	if headLength <= 0 {
		return nil, ErrorInvalidHeader
	}

	headerLength := codec.HeaderLength()
	buff := make([]byte, headerLength)
	n, err := reader.Read(buff)
	if err != nil {
		return nil, err
	}
	if n != headerLength {
		log.Warnf("read invalid header length: need %d, actual %d", headerLength, n)
	}

	header, err := codec.DecodeHeader(buff)
	if err != nil {
		return nil, err
	}

	bodyLength := header.BodyLength()
	if bodyLength < 0 || bodyLength >= limitSize {
		log.Warnf("invalid header length: %d", bodyLength)
		return nil, ErrorReadOutofRange
	}

	buff = make([]byte, bodyLength)
	n, err = reader.Read(buff)
	if err != nil {
		return nil, err
	}
	if n != bodyLength {
		log.Warnf("read invalid body length: need %d, actual %d", bodyLength, n)
	}

	body, err := codec.DecodeBody(header.Cmd(), buff)

	return &Message{
		Header: header,
		Body:   body,
	}, err
}

// 根据反射类型读取

// write 由服务端自己控制,不用限制字数
func WriteMessage(conn io.Writer, message *Message) error {
	return WriteLimitMessage(conn, message, 0) // unlimited
}

func WriteLimitMessage(conn io.Writer, message *Message, limitSize int) error {
	if message.Header == nil {
		return ErrorInvalidMessage
	}

	var body []byte
	var err error
	if message.Body != nil {
		body, err = message.Body.Encode()
		if err != nil {
			return err
		}
	}
	message.Header.SetBodyLength(len(body))

	if limitSize > 0 && len(body) > limitSize {
		return ErrorWriteOutofRange
	}

	buffer := bufPool.Get()
	defer bufPool.Put(buffer)

	hdr, err := message.Header.Encode()
	if err != nil {
		return err
	}
	buffer.Write(hdr)
	buffer.Write(body)
	_, err = conn.Write(buffer.Bytes())
	return err
}
