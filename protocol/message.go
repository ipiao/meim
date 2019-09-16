package protocol

import (
	"errors"
	"reflect"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/server"
)

var (
	ErrorInvalidMessage       = errors.New("invalid message")
	ErrorBodyLengthOutofRange = errors.New("body length out of range")
)

// 协议头
type ProrocolHeader interface {
	ProtocolData
	Length() int
	BodyLength() int
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
		return ErrorBodyLengthOutofRange
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
