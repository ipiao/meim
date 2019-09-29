package meim

// 由具体业务实现的方法函数
// must
type ExternalPlugin interface {
	HandleAuthClient(*Client) bool              // run的第一步, auth 认证客户端,至少确定协议方式,亦即 DataCreator
	HandleMessage(*Client, *Message)            // 消息处理函数
	HandleClientClosed(*Client)                 // 关闭客户端之后的处理
	HandleBeforeWriteMessage(*Client, *Message) //
}

//var (
//	ext     ExternalPlugin // ext = extension
//	extOnce sync.Once
//)
//
//// 只能调用一次
//func SetExternalPlugin(plugin ExternalPlugin) {
//	extOnce.Do(func() {
//		ext = plugin
//	})
//}
//
//// 客户端消息处理入队
//func HandleClientMessage(client *Client, msg *Message) {
//	client.EnqueueEvent(func(c *Client) {
//		ext.HandleMessage(c, msg)
//	})
//}
