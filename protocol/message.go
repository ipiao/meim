package protocol

import (
	"errors"
	"reflect"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/server"
	"github.com/ipiao/meim/util"
)

var (
	ErrorInvalidMessage  = errors.New("invalid message")
	ErrorReadOutofRange  = errors.New("read body length out of range")
	ErrorWriteOutofRange = errors.New("write body length out of range")

	bufPool = util.NewBufferPool()
)

// 协议头
type ProrocolHeader interface {
	ProtocolData
	Length() int
	BodyLength() int
	SetBodyLength(int)
}

// 有一个完整的消息包含头和body两部分
type Message struct {
	header ProrocolHeader
	body   ProtocolData
}

// 给定消息结构读取
func ReadLimitMessage(conn server.Conn, limitSize int, val *Message) error {
	if val.header == nil {
		return ErrorInvalidMessage
	}

	buff, err := conn.Read(val.header.Length())
	if err != nil {
		return err
	}

	err = val.header.Decode(buff)
	if err != nil {
		return err
	}

	bodyLength := val.header.BodyLength()
	if bodyLength < 0 || bodyLength >= limitSize {
		log.Warnf("invalid header length: %d", bodyLength)
		return ErrorReadOutofRange
	}

	if bodyLength > 0 && val.body == nil {
		return ErrorInvalidMessage
	}

	buff, err = conn.Read(bodyLength)
	if err != nil {
		return err
	}
	return val.body.Decode(buff)
}

// 根据反射类型读取
func ReadLimitMessage2(conn server.Conn, limitSize int, ht, bt reflect.Type) error {
	header := protocolDataPools.Get(ht).(ProrocolHeader)
	body := protocolDataPools.Get(bt).(ProtocolData)

	message := &Message{
		header: header,
		body:   body,
	}

	return ReadLimitMessage(conn, limitSize, message)
}

// write 由服务端自己控制,不用限制字数
func WriteMessage(conn server.Conn, message *Message) error {
	return WriteLimitMessage(conn, message, 0) // unlimited
}

func WriteLimitMessage(conn server.Conn, message *Message, limitSize int) error {
	if message.header == nil {
		return ErrorInvalidMessage
	}

	var body []byte
	var err error
	if message.body != nil {
		body, err = message.body.Encode()
		if err != nil {
			return err
		}
	}
	message.header.SetBodyLength(len(body))

	if limitSize > 0 && len(body) > limitSize {
		return ErrorWriteOutofRange
	}

	buffer := bufPool.Get()
	defer bufPool.Put(buffer)

	hdr, err := message.header.Encode()
	if err != nil {
		return err
	}
	buffer.Write(hdr)
	buffer.Write(body)
	_, err = conn.Write(buffer.Bytes())
	return err
}
