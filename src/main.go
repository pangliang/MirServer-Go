package main

import (
	"flag"
	"login_server"
)

func main() {
	flag.Parse()
	port := flag.Uint("port", 7000, "输入端口号")
	login_server.Start(*port)
}
