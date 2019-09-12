package comect

// 服务插件
type Plugin interface {
}

type (
	// 在Accept得到一个net.Conn后调用生成Conn
	// 函数需要达到阻塞运行的效果
	ConnAcceptPlugin interface {
		HandleConnAccept(Conn)
	}

	// 关闭连接以及关闭连接后的处理
	ConnClosePlugin interface {
		HandleConnClosed(Conn) // 关闭之后的处理
		HandleCloseConn(Conn)  // 主动关闭,需要在逻辑上保证连接会断开
	}
)

// 插件槽,通过实现插件幷注册进入插件槽,对流程某些环间进行控制
// pluginContainer implements PluginContainer interface.
type PluginContainer interface {
	ConnAcceptPlugin
	ConnClosePlugin
}

type pluginContainer struct {
	doHandleConnAccept func(Conn)
	doHandleConnClosed func(Conn)
	doHandleCloseConn  func(Conn)
}

func (pc *pluginContainer) SetPlugin(plugin Plugin) {
	if p, ok := plugin.(ConnAcceptPlugin); ok {
		pc.doHandleConnAccept = p.HandleConnAccept
	}

	if p, ok := plugin.(ConnClosePlugin); ok {
		pc.doHandleConnClosed = p.HandleConnClosed
		pc.doHandleCloseConn = p.HandleCloseConn
	}
}

func (pc *pluginContainer) HandleConnAccept(conn Conn) {
	if pc.doHandleConnAccept != nil {
		pc.doHandleConnAccept(conn)
	}
}

func (pc *pluginContainer) HandleConnClosed(conn Conn) {
	if pc.doHandleConnClosed != nil {
		pc.doHandleConnClosed(conn)
	}
}

func (pc *pluginContainer) HandleCloseConn(conn Conn) {
	if pc.doHandleCloseConn != nil {
		pc.doHandleCloseConn(conn)
	}
}
