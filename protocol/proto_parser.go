package protocol

import (
	"io"
)

type ProtoParser interface {
	ReadFrom(rr io.Reader) (p *Proto, err error)
	WriteTo(wr io.Writer, p *Proto) (err error)
}
