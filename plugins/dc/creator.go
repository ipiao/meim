package dc

import (
	"fmt"
	"reflect"

	"github.com/ipiao/meim"
	"github.com/ipiao/meim/log"
)

type DataCreator struct {
	headerType     reflect.Type
	cmdType        map[int]reflect.Type
	typeCmd        map[reflect.Type]int // 一般在写的时候需要
	cmdDescription map[int]string       // 消息描述信息,用于日志
	//mu         sync.RWMutex
}

func NewDataCreator() *DataCreator {
	return &DataCreator{
		cmdType:        make(map[int]reflect.Type),
		typeCmd:        make(map[reflect.Type]int),
		cmdDescription: make(map[int]string),
	}
}

func (m *DataCreator) SetHeaderType2(t reflect.Type) {
	m.headerType = t
}

func (m *DataCreator) SetHeaderType(header meim.ProtocolHeader) {
	m.SetHeaderType2(reflect.TypeOf(header))
}

func (m *DataCreator) SetBodyCmd2(cmd int, t reflect.Type, desc ...string) {
	if typ, ok := m.cmdType[cmd]; ok {
		log.Warnf("body cmd %d has been set type %s, will be replaced by %s", cmd, typ, t)
	}
	if c, ok := m.typeCmd[t]; ok {
		log.Warnf("body type %s has been set cmd %d, will be replaced by %d", t, c, cmd)
	}
	m.cmdType[cmd] = t
	m.typeCmd[t] = cmd
	if len(desc) > 0 {
		m.cmdDescription[cmd] = desc[0]
	}
}

func (m *DataCreator) SetBodyCmd(cmd int, body interface{}, desc ...string) {
	t := reflect.TypeOf(body)
	m.SetBodyCmd2(cmd, t, desc...)
}

func (m *DataCreator) GetCmd2(t reflect.Type) (int, bool) {
	cmd, ok := m.typeCmd[t]
	if !ok {
		log.Warnf("body %s doesnt set cmd", t)
	}

	return cmd, ok
}

func (m *DataCreator) GetCmd(body interface{}) (int, bool) {
	t := reflect.TypeOf(body)
	return m.GetCmd2(t)
}

func (m *DataCreator) GetDescription(cmd int) string {
	desc, ok := m.cmdDescription[cmd]
	if !ok {
		return fmt.Sprintf("CMD-%d", cmd)
	}
	return desc
}

func (m *DataCreator) GetMsg(cmd int) interface{} {
	t, ok := m.cmdType[cmd]
	if !ok {
		log.Warnf("cmd %d doesnt set body", cmd)
		return nil
	}
	return newTypeData(t)
}

func (m *DataCreator) Clone() *DataCreator {
	cts := make(map[int]reflect.Type)
	tcs := make(map[reflect.Type]int)
	cmdDescs := make(map[int]string)

	for cmd, bt := range m.cmdType {
		cts[cmd] = bt
	}

	for bt, cmd := range m.typeCmd {
		tcs[bt] = cmd
	}

	for cmd, desc := range m.cmdDescription {
		cmdDescs[cmd] = desc
	}

	return &DataCreator{
		headerType:     m.headerType,
		cmdType:        cts,
		typeCmd:        tcs,
		cmdDescription: cmdDescs,
	}
}

func (m *DataCreator) CreateHeader() meim.ProtocolHeader {
	return newTypeData(m.headerType).(meim.ProtocolHeader)
}

func (m *DataCreator) CreateBody(cmd int) meim.ProtocolBody {
	msg := m.GetMsg(cmd)
	if msg == nil {
		return nil
	}
	return msg.(meim.ProtocolBody)
}

func newTypeData(t reflect.Type) interface{} {
	var argv reflect.Value

	if t.Kind() == reflect.Ptr { // reply must be ptr
		argv = reflect.New(t.Elem())
	} else {
		argv = reflect.New(t)
	}

	return argv.Interface()
}

func (m *DataCreator) CreateMessage(body interface{}) *meim.Message {
	cmd, _ := m.GetCmd(body)
	hdr := m.CreateHeader()
	hdr.SetCmd(cmd)
	d, _ := body.(meim.ProtocolBody)
	return &meim.Message{
		Header: hdr,
		Body:   d,
	}
}
