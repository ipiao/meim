package gate

import (
	"container/list"
	"sync"

	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
	"go.uber.org/atomic"
)

type Connection struct {
	conn server.Conn

	closed    atomic.Bool            // 是否关闭
	mch       chan *protocol.Message // 一般消息下发通道, message channel
	lmsch     chan int               // 长消息下发(信号),long message signal channel
	lmessages *list.List             // 长消息存储队列
	extch     chan func(*Client)     // 外部时间队列, external event channel
	dc        protocol.DataCreator   //
	mu        sync.Mutex             // 锁

	uid      int64       // 用户id
	userdata interface{} // 用户数据
}
