package meim

import (
	"io"
	"net"

	"github.com/ipiao/meim.v2/log"
	"github.com/ipiao/meim.v2/protocol"
)

type PluginI interface {
	CreateHeader() protocol.ProtoHeader
	HandleMessage(client *Client, message *protocol.Message)
	HandleOnWriteError(client *Client, message *protocol.Message, err error) bool // 在写消息出错时候处理
	HandleBeforeWriteMessage(client *Client, message *protocol.Message)
}

var _ PluginI = DefaultPlugin()

// 必须实现
type Plugin struct {
	HeaderCreator      protocol.ProtoHeaderCreator
	MessageHandler     func(client *Client, message *protocol.Message)
	OnWriteError       func(client *Client, message *protocol.Message, err error) (shutdown bool)
	BeforeWriteMessage func(client *Client, message *protocol.Message)
}

func DefaultPlugin() *Plugin {
	return &Plugin{
		HeaderCreator: protocol.DefaultHeaderCreator,
		MessageHandler: func(client *Client, message *protocol.Message) {
			log.Infof("client %s, handle message %s", client.Log(), message.Log())
		},
		OnWriteError: func(client *Client, message *protocol.Message, err error) bool {
			if _, ok := err.(net.Error); ok || err == io.EOF { // maybe write on close chan
				log.Infof("[write-nil] client %s, msg : %s, err: %s", client.Log(), message, err)
			} else {
				log.Warnf("[write-err] client %s, msg : %s, err: %s", client.Log(), message, err)
			}
			return true
		},
		BeforeWriteMessage: func(client *Client, message *protocol.Message) {
			log.Infof("client %s, write message %s", client.Log(), message.Log())
		},
	}
}

func (p *Plugin) CreateHeader() protocol.ProtoHeader {
	if p.HeaderCreator != nil {
		return p.HeaderCreator()
	}
	return nil
}
func (p *Plugin) HandleMessage(client *Client, message *protocol.Message) {
	if p.MessageHandler != nil {
		p.MessageHandler(client, message)
	}
}
func (p *Plugin) HandleOnWriteError(client *Client, message *protocol.Message, err error) bool {
	if p.OnWriteError != nil {
		return p.OnWriteError(client, message, err)
	}
	return false
}
func (p *Plugin) HandleBeforeWriteMessage(client *Client, message *protocol.Message) {
	if p.BeforeWriteMessage != nil {
		p.BeforeWriteMessage(client, message)
	}
}
