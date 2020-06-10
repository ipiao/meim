package meim

import "fmt"

type plainData []byte

var (
	_ ProtocolHeader = &plainData{}
	_ ProtocolBody   = &plainData{}
)

func (d *plainData) Decode(b []byte) error {
	*d = b
	return nil
} // 从字节中读取
func (d *plainData) Encode() ([]byte, error) {
	return *d, nil
} // 编码
func (d *plainData) Length() int {
	return len(*d)
}
func (d *plainData) Cmd() int {
	return 0
}                               // 协议指令
func (d *plainData) SetCmd(int) {} // 指定协议指令
func (d *plainData) Seq() int   { return 0 }
func (d *plainData) SetSeq(int) {}
func (d *plainData) BodyLength() int {
	return 0
}
func (d *plainData) SetBodyLength(n int) {}
func (d *plainData) Ver() int {
	return 0
}
func (d *plainData) SetVer(v int) {}
func (d *plainData) Clone() ProtocolHeader {
	return d
}

func (d *plainData) String() string {
	return fmt.Sprintf("plain data, len: %d", len(*d))
}
