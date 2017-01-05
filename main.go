package main

import (
	"syscall"
	"log"
	"github.com/judwhite/go-svc/svc"
	"os"
	"path/filepath"
	"github.com/pangliang/MirServer-Go/loginserver"
	"github.com/pangliang/MirServer-Go/gameserver"
)

type program struct {
	loginServer *loginserver.LoginServer
	gameServer *gameserver.GameServer
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

	loginChan := make(chan interface{})

	p.loginServer = loginserver.New(loginChan)
	p.loginServer.Main()

	p.gameServer = gameserver.New(loginChan)
	p.gameServer.Main()

	return nil
}

func (p *program) Stop() error {
	if p.loginServer != nil {
		p.loginServer.Exit()
	}

	if p.gameServer != nil {
		p.gameServer.Exit()
	}
	return nil
}
