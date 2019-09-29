package meim

import (
	"errors"
	"fmt"
	"io"
	"reflect"

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

// 协议数据,定义了数据基本交换协议
// 头和body 都属于协议数据
type ProtocolData interface {
	Decode(b []byte) error   // 从字节中读取
	Encode() ([]byte, error) // 编码
}

// 协议头
type ProtocolHeader interface {
	ProtocolData
	Length() int
	Cmd() int   // 协议指令
	SetCmd(int) // 指定协议指令
	Seq() int
	SetSeq(int)
	BodyLength() int
	SetBodyLength(n int)
	Clone() ProtocolHeader
}

// 协议数据内容
type ProtocolBody = ProtocolData

// 一个完整的消息包含头和body两部分
type Message struct {
	Header ProtocolHeader
	Body   ProtocolBody
}

func (m *Message) String() string {
	return fmt.Sprintf("msg: %s, header: %v, body: %v", m.Header, m.Body)
}

// 协议数据创建器,可以分别创建头和body
// 定义DataCreator的作用之一是,在必要的时候,可以对不同的客户端使用不同的数据交换协议
type DataCreator interface {
	CreateHeader() ProtocolHeader
	CreateBody(cmd int) ProtocolBody
	GetCmd(body interface{}) (int, bool)
	GetCmd2(t reflect.Type) (int, bool)
	GetDescription(cmd int) string
}

// 不限制读
func ReadMessage(reader io.Reader, dc DataCreator) (*Message, error) {
	return ReadLimitMessage(reader, dc, 0)
}

// 限制读
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
	if body != nil {
		if bodyLength > 0 {
			buff = make([]byte, bodyLength)
			_, err = io.ReadFull(reader, buff)
			if err != nil {
				return nil, err
			}
			err = body.Decode(buff)
		}
	}
	return &Message{Header: header, Body: body}, err
}

// 解码字节流
func DecodeMessage(b []byte, dc DataCreator) (*Message, error) {
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

// 限制写
func WriteLimitMessage(conn io.Writer, message *Message, limitSize int) error {
	if message.Header == nil {
		return ErrorInvalidMessage
	}
	data, err := EncodeLimitMessage(message, limitSize)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

// 限制编码消息
func EncodeLimitMessage(message *Message, limitSize int) ([]byte, error) {
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
	if body != nil {
		buffer.Write(body)
	}
	return buffer.Bytes(), nil
}

// 编码Message
func EncodeMessage(message *Message) ([]byte, error) {
	return EncodeLimitMessage(message, 0)
}
