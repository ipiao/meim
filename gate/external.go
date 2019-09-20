package gate

import (
	"sync"

	"github.com/ipiao/meim/protocol"
)

// 由具体业务实现的方法函数
// must
type ExternalPlugin interface {
	HandleAuthClient(*Client)                 // run的第一步, auth 认证客户端,确定协议方式,亦即 DataCreator
	HandleMessage(*Client, *protocol.Message) // 处理后续消息
	HandleClientClosed(*Client)               // 关闭客户端之后的处理
}

var (
	exts     ExternalPlugin
	extsOnce sync.Once
)

// 只能调用一次
func SetExternalPlugin(plugin ExternalPlugin) {
	extsOnce.Do(func() {
		exts = plugin
	})
}
