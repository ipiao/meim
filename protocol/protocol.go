package protocol

// 协议数据
// 头和body 都属于协议数据
type ProtocolData interface {
	Decode(b []byte) error   // 从字节中读取
	Encode() ([]byte, error) // 编码
}

// 协议数据内容
type ProtocolBody = ProtocolData

// 协议头
type ProtocolHeader interface {
	ProtocolData
	Length() int
	Cmd() int   // 协议指令
	SetCmd(int) // 指定协议指令
	BodyLength() int
	SetBodyLength(n int)
}
