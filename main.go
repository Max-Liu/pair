package main

import (
	"pair/game"
	_ "pair/routers"

	"github.com/astaxie/beego"
	"github.com/googollee/go-socket.io"
)

func main() {
	go func() {

		gameServer := game.NewGameServer()
		chatFunc := func(so socketio.Socket, roomName string) func(msg string) {
			return func(msg string) {
				gameServer.Log.Informational("%s(%s) said %s", so.Id(), so.Request().RemoteAddr, msg)
				so.BroadcastTo(roomName, "chat message", msg)
			}
		}
		gameServer.RegisterEvent("chat message", chatFunc)
		gameServer.Start()
	}()
	beego.Run()
}
