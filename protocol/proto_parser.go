package protocol

import (
	"io"
)

type Parser interface {
	ReadFrom(rr io.Reader) (p *Proto, err error)
	WriteTo(wr io.Writer, p *Proto) (err error)
}

var ProtoParser Parser = &DefaultProtoParser{}

func ResetParser(p Parser) {
	ProtoParser = p
}

func ReadFrom(rr io.Reader) (p *Proto, err error) {
	return ProtoParser.ReadFrom(rr)
}
func WriteTo(wr io.Writer, p *Proto) (err error) {
	return ProtoParser.WriteTo(rr)
}
