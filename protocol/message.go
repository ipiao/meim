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
	Header ProrocolHeader
	Body   ProtocolBody
}

type DataCreator interface {
	CreateHeader() ProrocolHeader
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
	n, err := reader.Read(buff)
	if err != nil {
		return nil, err
	}
	if n != headerLength {
		log.Warnf("read invalid header length: need %d, actual %d", headerLength, n)
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
		n, err = reader.Read(buff)
		if err != nil {
			return nil, err
		}

		if n != bodyLength {
			log.Warnf("read invalid body length: need %d, actual %d", bodyLength, n)
		}

		err = body.Decode(buff)
	}
	return &Message{Header: header, Body: body}, err
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
