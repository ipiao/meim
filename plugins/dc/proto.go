package dc

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/ipiao/meim"
	"github.com/ipiao/meim/log"
)

// proto.Message
type ProtoData struct {
	proto.Message
}

func NewProtoData(data proto.Message) *ProtoData {
	return &ProtoData{data}
}

func (p *ProtoData) String() string {
	return p.Message.String()
}

func (p *ProtoData) Encode() ([]byte, error) {
	return proto.Marshal(p.Message)
}

func (p *ProtoData) Decode(b []byte) error {
	err := proto.Unmarshal(b, p.Message)
	if err != nil {
		log.Debugf("protodata decode err: %s", err)
	}
	return err
}

func (p *ProtoData) Reset() {
	p.Message.Reset()
}

// body 是proto.Message的创造器
type ProtoDataCreator struct {
	*DataCreator
}

func NewProtoDataCreator() *ProtoDataCreator {
	return &ProtoDataCreator{
		DataCreator: NewDataCreator(),
	}
}

// 必须是proto.Message类型的指针
func (m *ProtoDataCreator) SetBodyCmd2(cmd int, t reflect.Type, desc ...string) {
	if t.Kind() != reflect.Ptr || !t.Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
		log.Fatalf("invalid type of proto body: %s", t)
	}
	m.DataCreator.SetBodyCmd2(cmd, t, desc...)
}

//
func (m *ProtoDataCreator) SetBodyCmd(cmd int, body proto.Message, desc ...string) {
	t := reflect.TypeOf(body)
	m.SetBodyCmd2(cmd, t, desc...)
}

func (m *ProtoDataCreator) Clone() *ProtoDataCreator {
	return &ProtoDataCreator{
		DataCreator: m.DataCreator.Clone(),
	}
}

func (m *ProtoDataCreator) CreateBody(cmd int) meim.ProtocolBody {
	msg := m.GetMsg(cmd)
	if msg == nil {
		return nil
	}
	return NewProtoData(msg.(proto.Message))
}

func (m *ProtoDataCreator) CreateMessage(body interface{}) *meim.Message {
	cmd, _ := m.GetCmd(body)
	hdr := m.CreateHeader()
	hdr.SetCmd(cmd)
	d, _ := body.(proto.Message)
	return &meim.Message{
		Header: hdr,
		Body:   NewProtoData(d),
	}
}
