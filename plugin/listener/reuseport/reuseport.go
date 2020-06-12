package reuseport

import (
	"net"
	"regexp"
	"strings"

	meim "github.com/ipiao/meim"
	reuseport "github.com/kavu/go_reuseport"
)

func init() {
	meim.RegisterListenerMaker("reuseport", reuseportListenerMaker)
}

func reuseportListenerMaker(cfg *meim.Config) (ln net.Listener, err error) {
	var network string
	if validIP4(cfg.Address) {
		network = "tcp4"
	} else {
		network = "tcp6"
	}
	return reuseport.NewReusablePortListener(network, cfg.Address)
}

var ip4Reg = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	i := strings.LastIndex(ipAddress, ":")
	ipAddress = ipAddress[:i] //remove port

	return ip4Reg.MatchString(ipAddress)
}
