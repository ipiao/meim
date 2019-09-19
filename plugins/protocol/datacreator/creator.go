package datacreator

import (
	"reflect"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

type DataCreator struct {
	headerType reflect.Type
	bodyTypes  map[int]reflect.Type
	// mu sync.RWMutex
}

func NewDataCreator() *DataCreator {
	return &DataCreator{
		bodyTypes: make(map[int]reflect.Type),
	}
}

func (m *DataCreator) SetHeaderType(t reflect.Type) {
	m.headerType = t
}

func (m *DataCreator) SetBodyType(cmd int, t reflect.Type) {
	if typ, ok := m.bodyTypes[cmd]; ok {
		log.Warnf("body cmd %d has been set type %s, will be replaced by %s", cmd, typ, t)
	}
	m.bodyTypes[cmd] = t
}

func (m *DataCreator) Clone() *DataCreator {
	bts := make(map[int]reflect.Type)
	for cmd, bt := range m.bodyTypes {
		bts[cmd] = bt
	}
	return &DataCreator{
		headerType: m.headerType,
		bodyTypes:  bts,
	}
}

func (m *DataCreator) CreateHeader() protocol.ProrocolHeader {
	return newTypeData(m.headerType).(protocol.ProrocolHeader)
}

func (m *DataCreator) CreateBody(cmd int) protocol.ProtocolBody {
	t := m.bodyTypes[cmd]
	return newTypeData(t).(protocol.ProtocolBody)
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
