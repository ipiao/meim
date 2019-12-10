package protocol

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	dp := NewDefaultProtoParser()

	p := &Proto{
		Ver: 1,
		Op:  1,
	}
	wr := bytes.NewBuffer()
	dp.WriteTo(wr, p)
}
