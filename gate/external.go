package gate

import "github.com/ipiao/meim/protocol"

// 由具体业务实现的方法函数
// must
var (
	//
	HandleMessage    func(*Client, *protocol.Message)
	HandleAuthClient func(*Client) // auth 认证客户端,确定协议方式,亦即 DataCreator
	//
	HandleClientClosed func(*Client) // 关闭客户端之后的处理
)
