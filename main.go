package main

import (
	"flag"
	"github.com/pangliang/MirServer-Go/core"
	"github.com/pangliang/MirServer-Go/login"
)

func main() {
	flag.Parse()
	port := flag.Uint("port", 7000, "输入端口号")

	loginServer := core.CreateServer(uint32(*port), login.LoginHanders)
	loginServer.Start()
}
