package header

import "encoding/binary"

// 用的leaf header
type LeafHeader struct {
	From      uint8
	CodecType uint8
	BodyLen   uint32
	Command   uint32
	Version   uint16
	Seq       uint32
	Extra     uint64
}

func (h *LeafHeader) Decode(b []byte) error {
	h.From = b[0]
	h.CodecType = b[1]
	h.BodyLen = binary.LittleEndian.Uint32(b[2:6])
	h.Command = binary.LittleEndian.Uint32(b[6:10])
	h.Version = binary.LittleEndian.Uint16(b[10:12])
	h.Seq = binary.LittleEndian.Uint32(b[12:16])
	h.Extra = binary.LittleEndian.Uint64(b[16:24])
	return nil
}

func (h *LeafHeader) Encode() ([]byte, error) {
	b := make([]byte, 24)
	b[0] = h.From
	b[1] = h.CodecType
	binary.LittleEndian.PutUint32(b[2:6], h.BodyLen)
	binary.LittleEndian.PutUint32(b[6:10], h.Command)
	binary.LittleEndian.PutUint16(b[10:12], h.Version)
	binary.LittleEndian.PutUint32(b[12:16], h.Seq)
	binary.LittleEndian.PutUint64(b[16:24], h.Extra)
	return b, nil
}

func (h *LeafHeader) Length() int {
	return 24
}

func (h *LeafHeader) Cmd() int {
	return int(h.Command)
}

func (h *LeafHeader) SetCmd(cmd int) {
	h.Command = uint32(cmd)
}

func (h *LeafHeader) BodyLength() int {
	return int(h.BodyLen)
}

func (h *LeafHeader) SetBodyLength(n int) {
	h.BodyLen = uint32(n)
}
