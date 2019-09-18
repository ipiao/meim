package gate

import "github.com/ipiao/meim/protocol"

// 由具体业务实现的方法函数
// must
var (
	//
	HandleAuthClient func(*Client)                    // run的第一步, auth 认证客户端,确定协议方式,亦即 DataCreator
	HandleMessage    func(*Client, *protocol.Message) // 处理后续消息

	// HandleClientClosed func(*Client)                    // 关闭客户端之后的处理
)
