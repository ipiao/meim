package mars

import "encoding/binary"

// wx mars header
// 消息头部信息, 固定长度20
type MarsHeader struct {
	HeadLen uint32
	Version uint32
	Command uint32
	Seq     uint32
	BodyLen uint32
}

func (h *MarsHeader) Decode(b []byte) error {
	h.HeadLen = binary.BigEndian.Uint32(b[:4])
	h.Version = binary.BigEndian.Uint32(b[4:8])
	h.Command = binary.BigEndian.Uint32(b[8:12])
	h.Seq = binary.BigEndian.Uint32(b[12:16])
	h.BodyLen = binary.BigEndian.Uint32(b[16:20])
	return nil
}

func (h *MarsHeader) Encode() ([]byte, error) {
	b := make([]byte, 20, 20)
	binary.BigEndian.PutUint32(b[:4], h.HeadLen)
	binary.BigEndian.PutUint32(b[4:8], h.Version)
	binary.BigEndian.PutUint32(b[8:12], h.Command)
	binary.BigEndian.PutUint32(b[12:16], h.Seq)
	binary.BigEndian.PutUint32(b[16:20], h.BodyLen)
	return b, nil
}

func (h *MarsHeader) Length() int {
	return 20
}

func (h *MarsHeader) Cmd() int {
	return int(h.Command)
}

func (h *MarsHeader) SetCmd(cmd int) {
	h.Command = uint32(cmd)
}

func (h *MarsHeader) BodyLength() int {
	return int(h.BodyLen)
}

func (h *MarsHeader) SetBodyLength(n int) {
	h.BodyLen = uint32(n)
}
