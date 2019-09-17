package datacreator

import "github.com/golang/protobuf/proto"

type ProtoData struct {
	proto.Message
}

func NewProtoData(data proto.Message) *ProtoData {
	return &ProtoData{data}
}

func (p *ProtoData) Encode() ([]byte, error) {
	return proto.Marshal(p.Message)
}

func (p *ProtoData) Decode(b []byte) error {
	return proto.Unmarshal(b, p.Message)
}

func (p *ProtoData) Reset() {
	p.Message.Reset()
}
