package meim

import (
	"sync"

	"github.com/ipiao/meim/server"
	"go.uber.org/atomic"
)

type Client struct {
	server.Channel
	PluginI
	cfg ClientConfig
	// 使用chan提供简单的缓冲、同步功能（在业务中如果做了相应的功能就可以直接使用WriteMessage）
	mu     sync.Mutex  // 锁
	closed atomic.Bool // 是否关闭

	UID      int64       // 用户id
	Version  int16       // 版本
	UserData interface{} // 用户其他私有数据
}

func NewClient() {

}
