package mars

import "encoding/binary"

// wx mars header
// 消息头部信息, 固定长度20
type Header struct {
	HeadLen uint32
	Version uint32
	Cmd     uint32
	Seq     uint32
	BodyLen uint32
}

func (h *Header) Decode(b []byte) error {
	h.HeadLen = binary.BigEndian.Uint32(b[:4])
	h.Version = binary.BigEndian.Uint32(b[4:8])
	h.Cmd = binary.BigEndian.Uint32(b[8:12])
	h.Seq = binary.BigEndian.Uint32(b[12:16])
	h.BodyLen = binary.BigEndian.Uint32(b[16:20])
	return nil
}

func (h *Header) Encode() ([]byte, error) {
	b := make([]byte, 20, 20)
	binary.BigEndian.PutUint32(b[:4], h.HeadLen)
	binary.BigEndian.PutUint32(b[4:8], h.Version)
	binary.BigEndian.PutUint32(b[8:12], h.Cmd)
	binary.BigEndian.PutUint32(b[12:16], h.Seq)
	binary.BigEndian.PutUint32(b[16:20], h.BodyLen)
	return b, nil
}

func (h *Header) Length() int {
	return 20
}

func (h *Header) BodyLength() int {
	return int(h.BodyLen)
}

func (h *Header) SetBodyLength(n int) {
	h.BodyLen = uint32(n)
}
