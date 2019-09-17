package protocol

// 协议数据
// 头和body 都属于协议数据
type ProtocolData interface {
	Decode(b []byte) error   // 从字节中读取
	Encode() ([]byte, error) // 编码
}

// 协议数据内容
type ProtocolBody = ProtocolData
