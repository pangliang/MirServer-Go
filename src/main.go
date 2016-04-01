package main

import (
	"flag"
	"login"
	"core"
)

func main() {
	flag.Parse()
	port := flag.Uint("port", 7000, "输入端口号")

	loginServer := core.Server{uint32(*port), login.LoginHanders}
	loginServer.Start()
}
