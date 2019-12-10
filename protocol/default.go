package protocol

import (
	"errors"
	"io"

	"github.com/ipiao/meim.v2/libs/encoding/binary"
)

var (
	ErrProtoBodyLen   = errors.New("invalid proto body length")
	ErrProtoHeaderLen = errors.New("invalid proto header length")
)

const (
	// size
	_headerSize      = 2 // 头长，这样header和body可以不必连在一起
	_verSize         = 2 // 版本号字长
	_opSize          = 4 // 指令字长
	_seqSize         = 4 // 序列号字长
	_compressSize    = 1 // 压缩类型字长
	_contentTypeSize = 1 // body内容类型字长
	_bodyLenSize     = 4 // body长

	_rawHeaderSize = _headerSize + _verSize + _opSize + _seqSize + _compressSize + _contentTypeSize + _bodyLenSize
	// offset
	_headerOffset      = 0
	_verOffset         = _headerOffset + _headerSize
	_opOffset          = _verOffset + _verSize
	_seqOffset         = _opOffset + _opSize
	_compressOffset    = _seqOffset + _seqSize
	_contentTypeOffset = _compressOffset + _compressSize
	_bodyLenOffset     = _contentTypeOffset + _contentTypeSize
)

type DefaultProtoParser struct {
	MaxBodySize int // 1<< 12
	//pool        *buffer.Pool
}

var _ ProtoParser = &DefaultProtoParser{}

// 默认解析器
func NewDefaultProtoParser() *DefaultProtoParser {
	return &DefaultProtoParser{
		MaxBodySize: 1 << 12,
		//pool:        buffer.New(1024),
	} // 4k
}

func (dp *DefaultProtoParser) ReadFrom(rr io.Reader) (p *Proto, err error) {
	var (
		headerLen int16
		bodyLen   int
		buf       = make([]byte, _rawHeaderSize)
	)
	if _, err = io.ReadFull(rr, buf); err != nil {
		return
	}
	p = new(Proto)
	headerLen = binary.BigEndian.Int16(buf[_headerOffset:])
	p.Ver = int32(binary.BigEndian.Int16(buf[_verOffset:]))
	p.Op = binary.BigEndian.Int32(buf[_opOffset:])
	p.Seq = binary.BigEndian.Int32(buf[_seqOffset:])
	p.Compress = int32(buf[_compressOffset])
	p.ContentType = int32(buf[_contentTypeOffset])
	bodyLen = int(binary.BigEndian.Int32(buf[_bodyLenOffset:]))
	if dp.MaxBodySize > 0 && bodyLen > dp.MaxBodySize {
		err = ErrProtoBodyLen
		return
	}
	if headerLen != _rawHeaderSize {
		err = ErrProtoHeaderLen
		return
	}
	if bodyLen > 0 {
		buf = make([]byte, bodyLen)
		_, err = io.ReadFull(rr, buf)
		p.Body = buf
	} else {
		p.Body = nil
	}
	return
}

func (dp *DefaultProtoParser) WriteTo(wr io.Writer, p *Proto) (err error) {
	var (
		bodyLen = int32(len(p.Body))
		buf     = make([]byte, _rawHeaderSize)
	)
	binary.BigEndian.PutInt16(buf[_headerOffset:], _rawHeaderSize)
	binary.BigEndian.PutInt16(buf[_verOffset:], int16(p.Ver))
	binary.BigEndian.PutInt32(buf[_opOffset:], p.Op)
	binary.BigEndian.PutInt32(buf[_seqOffset:], p.Seq)
	buf[_compressOffset] = byte(p.Compress)
	buf[_contentTypeOffset] = byte(p.ContentType)
	binary.BigEndian.PutInt32(buf[_bodyLenOffset:], bodyLen)

	if p.Body != nil {
		newBuf := make([]byte, _rawHeaderSize+bodyLen)
		copy(newBuf[:_rawHeaderSize], buf)
		copy(newBuf[_rawHeaderSize:], p.Body)
		buf = newBuf
	}
	_, err = wr.Write(buf)
	return
}
