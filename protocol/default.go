package protocol

import (
	"encoding/binary"
	"errors"
)

var (
	ErrorInvalidHeaderBuffLen     = errors.New("invalid header buff length")
	ErrorInvalidHeaderProtocolLen = errors.New("invalid header protocol length")
)

const (
	// size
	_headerSize      = 2 // 头长，这样header和body可以不必连在一起
	_verSize         = 2 // 版本号字长
	_cmdSize         = 4 // 指令字长
	_seqSize         = 4 // 序列号字长
	_compressSize    = 1 // 压缩类型字长
	_contentTypeSize = 1 // body内容类型字长
	_bodySize        = 4 // body长

	_rawHeaderSize = _headerSize + _verSize + _cmdSize + _seqSize + _compressSize + _contentTypeSize + _bodySize
	// offset
	_headerOffset      = 0
	_verOffset         = _headerOffset + _headerSize
	_cmdOffset         = _verOffset + _verSize
	_seqOffset         = _cmdOffset + _cmdSize
	_compressOffset    = _seqOffset + _seqSize
	_contentTypeOffset = _compressOffset + _compressSize
	_bodyOffset        = _contentTypeOffset + _contentTypeSize
)

var DefaultHeaderCreator = func() ProtoHeader {
	return new(DefaultHeader)
}

type DefaultHeader struct {
	ver         int16
	cmd         int32
	seq         int32
	compress    int8
	contentType int8
	bodyLen     int32
}

func (h *DefaultHeader) Len() int                        { return _rawHeaderSize }
func (h *DefaultHeader) Ver() int16                      { return h.ver }
func (h *DefaultHeader) SetVer(ver int16)                { h.ver = ver }
func (h *DefaultHeader) Cmd() int32                      { return h.cmd }
func (h *DefaultHeader) SetCmd(cmd int32)                { h.cmd = cmd }
func (h *DefaultHeader) Seq() int32                      { return h.seq }
func (h *DefaultHeader) SetSeq(seq int32)                { h.seq = seq }
func (h *DefaultHeader) Compress() int8                  { return h.compress }
func (h *DefaultHeader) SetCompress(compress int8)       { h.compress = compress }
func (h *DefaultHeader) ContentType() int8               { return h.contentType }
func (h *DefaultHeader) SetContentType(contentType int8) { h.contentType = contentType }
func (h *DefaultHeader) BodyLen() int                    { return int(h.bodyLen) }
func (h *DefaultHeader) SetBodyLen(bodyLen int)          { h.bodyLen = int32(bodyLen) }
func (h *DefaultHeader) ToData() []byte {
	buf := make([]byte, h.Len(), h.Len())
	binary.BigEndian.PutUint16(buf[_headerOffset:], uint16(_rawHeaderSize))
	binary.BigEndian.PutUint16(buf[_verOffset:], uint16(h.ver))
	binary.BigEndian.PutUint32(buf[_cmdOffset:], uint32(h.cmd))
	binary.BigEndian.PutUint32(buf[_seqOffset:], uint32(h.seq))
	buf[_compressOffset] = byte(h.compress)
	buf[_contentTypeOffset] = byte(h.contentType)
	binary.BigEndian.PutUint32(buf[_bodyOffset:], uint32(h.bodyLen))
	return buf
}
func (h *DefaultHeader) FromData(buf []byte) error {
	if len(buf) < h.Len() {
		return ErrorInvalidHeaderBuffLen
	}
	headerLen := int(binary.BigEndian.Uint16(buf[_headerOffset:_verOffset]))
	if headerLen != h.Len() {
		return ErrorInvalidHeaderProtocolLen
	}
	h.ver = int16(binary.BigEndian.Uint16(buf[_verOffset:_cmdOffset]))
	h.cmd = int32(binary.BigEndian.Uint32(buf[_cmdOffset:_seqOffset]))
	h.seq = int32(binary.BigEndian.Uint32(buf[_seqOffset:_compressOffset]))
	h.compress = int8(buf[_compressOffset])
	h.contentType = int8(buf[_contentTypeOffset])
	h.bodyLen = int32(binary.BigEndian.Uint32(buf[_bodyOffset : _bodyOffset+_bodySize]))
	return nil
}

var _ ProtoHeader = new(DefaultHeader)

type DefaultBody []byte

func (b *DefaultBody) ToData() []byte {
	return *b
}
func (b *DefaultBody) FromData(buf []byte) error {
	*b = buf
	return nil
}
