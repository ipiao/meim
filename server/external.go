package server

import "errors"

var (
	HandleConnAccepted func(Conn)
	HandleCloseConn    func(Conn)
	HandleConnClosed   func(Conn)
)

//
func CheckExternalHandlers() error {
	if HandleConnAccepted == nil {
		return errors.New("external handler HandleConnAccepted not set")
	}

	if HandleCloseConn == nil {
		return errors.New("external handler HandleCloseConn not set")
	}

	if HandleConnClosed == nil {
		return errors.New("external handler HandleConnClosed not set")
	}
	return nil
}

// // 全局插件
// var Plugins PluginContainer

// func init() {
// 	Plugins = &pluginContainer{}
// }

// // 服务插件
// type Plugin interface{}

// type (
// 	// 在Accept得到一个net.Conn后调用生成Conn
// 	// 函数需要达到阻塞运行的效果
// 	ConnAcceptedPlugin interface {
// 		HandleConnAccepted(Conn)
// 	}

// 	// 关闭连接以及关闭连接后的处理
// 	ConnClosePlugin interface {
// 		HandleConnClosed(Conn) // 关闭之后的处理
// 		HandleCloseConn(Conn)  // 主动关闭,需要在逻辑上保证连接会断开
// 	}
// )

// // 插件槽,通过实现插件幷注册进入插件槽,对流程某些环间进行控制
// // pluginContainer implements PluginContainer interface.
// type PluginContainer interface {
// 	Check() error            // check 检查插件功能的完备性,有些插件是必须实现的
// 	SetPlugin(plugin Plugin) // 实现插件注册
// 	ConnAcceptedPlugin
// 	ConnClosePlugin
// }

// // 需要外部调用者注入实现的方法函数
// type pluginContainer struct {
// 	doHandleConnAccepted func(Conn)
// 	doHandleConnClosed   func(Conn)
// 	doHandleCloseConn    func(Conn)
// }

// func (pc *pluginContainer) SetPlugin(plugin Plugin) {
// 	if p, ok := plugin.(ConnAcceptedPlugin); ok {
// 		pc.doHandleConnAccepted = p.HandleConnAccepted
// 	}

// 	if p, ok := plugin.(ConnClosePlugin); ok {
// 		pc.doHandleConnClosed = p.HandleConnClosed
// 		pc.doHandleCloseConn = p.HandleCloseConn
// 	}

// }

// //
// func (pc *pluginContainer) HandleConnAccepted(conn Conn) {
// 	if pc.doHandleConnAccepted != nil {
// 		pc.doHandleConnAccepted(conn)
// 	}
// }

// //
// func (pc *pluginContainer) HandleConnClosed(conn Conn) {
// 	if pc.doHandleConnClosed != nil {
// 		pc.doHandleConnClosed(conn)
// 	}
// }

// // must
// func (pc *pluginContainer) HandleCloseConn(conn Conn) {
// 	if pc.doHandleCloseConn != nil {
// 		pc.doHandleCloseConn(conn)
// 	}
// }
