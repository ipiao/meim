package protocol

import (
	"errors"
	"fmt"
	"io"

	"github.com/ipiao/meim/log"
)

var (
	ErrorReadOutOfRange  = errors.New("read body length out of range")
	ErrorWriteOutOfRange = errors.New("write body length out of range")
	ErrorInvalidMessage  = errors.New("invalid message")

	DefaultReadLimit  = 1 << 12
	DefaultWriteLimit = 1 << 22
)

type Message struct {
	Header ProtoHeader
	Body   []byte
}

func (m *Message) Log() string {
	if m.Header != nil {
		return logHeader(m.Header)
	}
	return "nil message"
}

// header 创建器
type ProtoHeaderCreator func() ProtoHeader

// 数据头
// 为了在进行框架切换的时候版本兼容，使用interface Header
type ProtoHeader interface {
	Len() int                   // 头自身长度，推荐固定长
	Ver() int16                 // Version,版本号
	SetVer(int16)               //
	Cmd() int32                 // Command,指令号，固定指令(建议固定指令，由框架处理，业务指令放到body)
	SetCmd(int32)               //
	Seq() int32                 // Sequence,客户端序列号
	SetSeq(int32)               //
	Compress() int8             // 压缩类型
	SetCompress(int8)           //
	ContentType() int8          // 内容类型/编码类型
	SetContentType(int8)        //
	BodyLen() int               // body长
	SetBodyLen(int)             // used when write
	ToData() []byte             //
	FromData(buff []byte) error //
}

func logHeader(h ProtoHeader) string {
	return fmt.Sprintf("ver: %d, cmd: %d, seq: %d, compress: %d, contentType: %d, bodyLen: %d",
		h.Ver(), h.Cmd(), h.Seq(), h.Compress(), h.ContentType(), h.BodyLen())
}

// 不限制读
func ReadMessage(reader io.Reader, message *Message) error {
	return ReadLimitMessage(reader, message, DefaultReadLimit)
}

// 限制读
func ReadLimitMessage(reader io.Reader, message *Message, limitSize int) error {
	if message.Header == nil {
		return ErrorInvalidMessage
	}

	buff := make([]byte, message.Header.Len())
	_, err := io.ReadFull(reader, buff)
	if err != nil {
		return err
	}

	err = message.Header.FromData(buff)
	if err != nil {
		return err
	}

	bodyLength := message.Header.BodyLen()
	if bodyLength < 0 || (limitSize > 0 && int(bodyLength) > limitSize) {
		log.Warnf("invalid body length: %d", bodyLength)
		return ErrorReadOutOfRange
	}
	if bodyLength > 0 {
		buff = make([]byte, bodyLength)
		_, err = io.ReadFull(reader, buff)
		if err != nil {
			return err
		}
		message.Body = buff
	}
	return err
}

// 写
func WriteMessage(writer io.Writer, message *Message, limitSize int) error {
	return WriteLimitMessage(writer, message, DefaultWriteLimit)
}

// 限制写
func WriteLimitMessage(writer io.Writer, message *Message, limitSize int) error {
	if message.Header == nil {
		return ErrorInvalidMessage
	}
	bodyLength := len(message.Body)
	if bodyLength > 0 && bodyLength > limitSize {
		log.Warnf("invalid body length: %d", bodyLength)
		return ErrorWriteOutOfRange
	}
	message.Header.SetBodyLen(bodyLength)
	_, err := writer.Write(message.Header.ToData())
	if err != nil {
		return err
	}
	if bodyLength > 0 {
		_, err = writer.Write(message.Body)
	}
	return err
}
