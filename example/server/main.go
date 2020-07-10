package main

import (
	"flag"

	"github.com/ipiao/meim"
	"github.com/ipiao/meim/conf"
)

func main() {
	flag.Parse()
	s := meim.NewServer(conf.Conf)
	s.Run()
	select {}
}
