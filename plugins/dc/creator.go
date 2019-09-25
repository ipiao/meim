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

func (m *DataCreator) SetHeaderType(t reflect.Type) {
	m.headerType = t
}

func (m *DataCreator) SetHeaderType2(header meim.ProtocolHeader) {
	m.SetHeaderType(reflect.TypeOf(header))
}

func (m *DataCreator) SetBodyCmd(cmd int, t reflect.Type, desc ...string) {
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

func (m *DataCreator) SetBodyCmd2(cmd int, body meim.ProtocolBody, desc ...string) {
	t := reflect.TypeOf(body)
	m.SetBodyCmd(cmd, t, desc...)
}

func (m *DataCreator) GetCmd(t reflect.Type) (int, bool) {
	cmd, ok := m.typeCmd[t]
	if !ok {
		log.Warnf("body %s doesnt set cmd", t)
	}
	return cmd, ok
}

func (m *DataCreator) GetCmd2(body interface{}) (int, bool) {
	t := reflect.TypeOf(body)
	return m.GetCmd(t)
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

	for cmd, bt := range m.cmdType {
		cts[cmd] = bt
		tcs[bt] = cmd
	}

	return &DataCreator{
		headerType: m.headerType,
		cmdType:    cts,
		typeCmd:    tcs,
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
