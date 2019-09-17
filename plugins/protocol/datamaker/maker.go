package datamaker

import (
	"reflect"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

type DataMaker struct {
	headerType reflect.Type
	bodyTypes  map[int]reflect.Type
	// typeCmd map[reflect.Type]int
	// mu sync.RWMutex
}

func NewDataMaker() *DataMaker {
	return &DataMaker{
		bodyTypes: make(map[int]reflect.Type),
	}
}

func (m *DataMaker) SetHeaderType(t reflect.Type) {
	m.headerType = t
}

func (m *DataMaker) SetBodyType(cmd int, t reflect.Type) {
	if typ, ok := m.bodyTypes[cmd]; ok {
		log.Warnf("body cmd %d has been set type %s, will be replaced by %s", cmd, typ, t)
	}
	m.bodyTypes[cmd] = t
}

func (m *DataMaker) Clone() *DataMaker {
	bts := make(map[int]reflect.Type)
	for cmd, bt := range m.bodyTypes {
		bts[cmd] = bt
	}
	return &DataMaker{
		headerType: m.headerType,
		bodyTypes:  bts,
	}
}

func (m *DataMaker) MakeHeader() protocol.ProrocolHeader {
	return newTypeData(m.headerType).(protocol.ProrocolHeader)
}

func (m *DataMaker) MakeBody(cmd int) protocol.ProtocolBody {
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
