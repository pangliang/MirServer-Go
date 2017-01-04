package main

import (
	"flag"
	"syscall"
	"log"
	"github.com/judwhite/go-svc/svc"
	"os"
	"path/filepath"
	"github.com/pangliang/MirServer-Go/util"
	"net"
	"github.com/pangliang/MirServer-Go/login"
	"github.com/pangliang/MirServer-Go/protocol"
)

type program struct {
	loginServerListener net.Listener
	waitGroup   util.WaitGroupWrapper
}

func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}


func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

func (p *program) Start() error {
	flagSet := flag.NewFlagSet("mirserver", flag.ExitOnError)
	loginAddress := flagSet.String("login-address", "0.0.0.0:7000", "<addr>:<port> to listen on for TCP clients")

	flagSet.Parse(os.Args[1:])

	listener, err := net.Listen("tcp", *loginAddress)
	if err != nil {
		log.Fatalln("start server error: ", err)
	}

	p.loginServerListener = listener

	loginServer := login.New()

	p.waitGroup.Wrap(func() {
		protocol.TCPServer(listener, loginServer)
	})
	return nil
}

func (p *program) Stop() error {
	if p.loginServerListener != nil {
		p.loginServerListener.Close()
	}
	return nil
}
