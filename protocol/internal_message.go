package protocol

import (
	"encoding/binary"
	"io"
)

// 内部消息
type InternalMessage struct {
	Message         // 发送的消息体
	Sender    int64 // 发送人
	Receiver  int64 // 接收人
	Timestamp int64 // 时间戳,ms
}

func WriteInternalMessage(conn io.Writer, msg *InternalMessage) error {
	data, err := MarshalInternalMessage(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

func MarshalInternalMessage(message *InternalMessage) ([]byte, error) {
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

	message.Header.SetBodyLength(len(body) + 24)

	buffer := bufPool.Get()
	defer bufPool.Put(buffer)

	hdr, err := message.Header.Encode()
	if err != nil {
		return nil, err
	}

	buffer.Write(hdr)
	binary.Write(buffer, binary.BigEndian, message.Sender)
	binary.Write(buffer, binary.BigEndian, message.Receiver)
	binary.Write(buffer, binary.BigEndian, message.Timestamp)
	buffer.Write(body)
	return buffer.Bytes(), nil
}

// 编码Message
func UnmarshalInternalMessgae(b []byte, dc DataCreator) (*InternalMessage, error) {
	message := new(InternalMessage)
	message.Header = dc.CreateHeader()

	headerLength := message.Header.Length()

	if len(b) < headerLength {
		return message, ErrorReadOutofRange
	}
	head := b[:headerLength]
	err := message.Header.Decode(head)
	if err != nil {
		return message, err
	}

	bodyLength := message.Header.BodyLength()
	if bodyLength != len(b)-headerLength {
		return message, ErrorInvalidMessage
	}

	message.Body = dc.CreateBody(message.Header.Cmd())
	if bodyLength >= 24 {
		message.Sender = int64(binary.BigEndian.Uint64(b[headerLength : headerLength+8]))
		message.Receiver = int64(binary.BigEndian.Uint64(b[headerLength+8 : headerLength+16]))
		message.Timestamp = int64(binary.BigEndian.Uint64(b[headerLength+16 : headerLength+24]))

		err = message.Body.Decode(b[headerLength+24:])
	} else {
		return message, ErrorInvalidMessage
	}

	return message, err
}

// 编码Message
func ReadInternalMessage(reader io.Reader, dc DataCreator) (*InternalMessage, error) {
	message := new(InternalMessage)
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

	message.Header = header

	bodyLength := header.BodyLength()

	body := dc.CreateBody(header.Cmd())
	if bodyLength >= 24 {
		buff = make([]byte, bodyLength)
		_, err = io.ReadFull(reader, buff)
		if err != nil {
			return nil, err
		}

		message.Sender = int64(binary.BigEndian.Uint64(buff[:8]))
		message.Receiver = int64(binary.BigEndian.Uint64(buff[8:16]))
		message.Timestamp = int64(binary.BigEndian.Uint64(buff[16:24]))

		err = body.Decode(buff[24:])
	} else {
		return message, ErrorInvalidMessage
	}
	message.Body = body
	return message, err
}
