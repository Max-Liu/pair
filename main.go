package main

import (
	"pair/game"
	_ "pair/routers"

	"github.com/astaxie/beego"
)

func main() {
	go game.NewGameServer().Start()
	beego.Run()
}
