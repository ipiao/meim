package comect

// 服务插件
type Plugin interface {
}

type (
	// PostConnAcceptPlugin 在Accept得到一个net.Conn后调用生成Conn
	PostConnAcceptPlugin interface {
		HandleConnAccept(Conn) Conn
	}
)

// 插件槽,通过实现插件幷注册进入插件槽,对流程某些环间进行控制
// pluginContainer implements PluginContainer interface.
type PluginContainer interface {
	PostConnAcceptPlugin
}

type pluginContainer struct {
	doPostConnAccept func(Conn) Conn
}

func (pc *pluginContainer) SetPlugin(plugin Plugin) {
	if p, ok := plugin.(PostConnAcceptPlugin); ok {
		pc.doPostConnAccept = p.HandleConnAccept
	}
}

func (pc *pluginContainer) WrapPlugin(plugin Plugin) {

	if p, ok := plugin.(PostConnAcceptPlugin); ok {
		if pc.doPostConnAccept != nil {
			pc.doPostConnAccept = func(conn Conn) Conn {
				conn = pc.doPostConnAccept(conn)
				return p.HandleConnAccept(conn)
			}
		} else {
			pc.doPostConnAccept = p.HandleConnAccept
		}
	}

}

func (pc *pluginContainer) HandleConnAccept(conn Conn) Conn {
	if pc.doPostConnAccept != nil {
		return pc.doPostConnAccept(conn)
	}
	return conn
}
