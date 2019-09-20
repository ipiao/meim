package server

import "sync"

type ExternalPlugin interface {
	HandleConnAccepted(Conn)
	HandleCloseConn(Conn)
	HandleConnClosed(Conn)
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
