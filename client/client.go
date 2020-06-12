package client

import (
	"net"

	meim "github.com/ipiao/meim"
)

// TODO: dialer for listener
func DialTCP(addr string, config meim.ClientConfig, plugin meim.PluginI) (*meim.Client, error) {
	config.Init()
	if plugin == nil {
		plugin = meim.DefaultPlugin()
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	netConn := meim.NewNetConn(conn, 0, config.WriteTimeout)
	client := meim.NewClient(netConn, config, plugin)
	return client, nil
}
