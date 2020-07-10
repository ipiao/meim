package main

import (
	"flag"

	"github.com/ipiao/meim/conf"

	"github.com/ipiao/meim"
)

func main() {
	flag.Parse()
	s := meim.NewServer(conf.Conf)
	s.Run()
	select {}
}
