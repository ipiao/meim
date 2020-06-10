package header

import (
	"encoding/binary"
	"fmt"

	"github.com/ipiao/meim"
)

var (
	_ meim.ProtocolBody = &MarsHeader{}
)

// wx mars header
// 消息头部信息, 固定长度20
type MarsHeader struct {
	HeadLen  uint32
	Version  uint32
	Command  uint32
	Sequence uint32
	BodyLen  uint32
}

func (h *MarsHeader) String() string {
	return fmt.Sprintf("cmd: %d, seq %d, bodylen %d", h.Command, h.Sequence, h.BodyLen)
}

func (h *MarsHeader) Decode(b []byte) error {
	h.HeadLen = binary.BigEndian.Uint32(b[:4])
	h.Version = binary.BigEndian.Uint32(b[4:8])
	h.Command = binary.BigEndian.Uint32(b[8:12])
	h.Sequence = binary.BigEndian.Uint32(b[12:16])
	h.BodyLen = binary.BigEndian.Uint32(b[16:20])
	return nil
}

func (h *MarsHeader) Encode() ([]byte, error) {
	b := make([]byte, 20, 20)
	h.HeadLen = 20
	binary.BigEndian.PutUint32(b[:4], h.HeadLen)
	binary.BigEndian.PutUint32(b[4:8], h.Version)
	binary.BigEndian.PutUint32(b[8:12], h.Command)
	binary.BigEndian.PutUint32(b[12:16], h.Sequence)
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

func (h *MarsHeader) SetSeq(seq int) {
	h.Sequence = uint32(seq)
}

func (h *MarsHeader) Seq() int {
	return int(h.Sequence)
}

func (h *MarsHeader) Ver() int {
	return int(h.Version)
}

func (h *MarsHeader) SetVer(v int) {
	h.Version = uint32(v)
}

func (h *MarsHeader) BodyLength() int {
	return int(h.BodyLen)
}

func (h *MarsHeader) SetBodyLength(n int) {
	h.BodyLen = uint32(n)
}

func (h *MarsHeader) Clone() meim.ProtocolHeader {
	return &MarsHeader{
		HeadLen:  h.HeadLen,
		Version:  h.Version,
		Command:  h.Command,
		Sequence: h.Sequence,
		BodyLen:  h.BodyLen,
	}
}
