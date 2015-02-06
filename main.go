package main

import (
	"encoding/json"
	"pair/game"
	_ "pair/routers"
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/googollee/go-socket.io"
)

func main() {
	go func() {

		var confirm = struct {
			sync.Mutex
			Map map[string][]string
		}{Map: make(map[string][]string)}

		var gameLog = struct {
			sync.Mutex
			Map map[string]map[string]string
		}{Map: make(map[string]map[string]string)}

		var gameLogTrue = struct {
			sync.Mutex
			Map map[string][]string
		}{Map: make(map[string][]string)}

		gameServer := game.NewGameServer()
		chatFunc := func(so socketio.Socket, roomName string) func(msg string) {
			return func(msg string) {
				gameServer.Log.Informational("%s(%s) said %s", so.Id(), so.Request().RemoteAddr, msg)
				so.BroadcastTo(roomName, "chatmsg", msg)
			}
		}

		confirmFunc := func(so socketio.Socket, roomName string) func(msg string) {
			return func(msg string) {
				gameServer.Log.Informational("%s(%s) confirmed.", so.Id(), so.Request().RemoteAddr)
				confirm.Lock()
				defer confirm.Unlock()
				peopleSlice, ok := confirm.Map[roomName]
				if ok {
					if len(peopleSlice) == 2 {
						if peopleSlice[0] != so.Id() {
							peopleSlice[1] = so.Id()

							gameServer.Log.Informational("%s(%s) Game(%s) started.", so.Id(), so.Request().RemoteAddr, roomName)
							delete(confirm.Map, roomName)
							gameServer.BroadcastTo(roomName, "gamestart")
						}
					}
					if len(peopleSlice) == 1 {
						delete(confirm.Map, roomName)
					}
					if len(peopleSlice) == 0 {
						delete(confirm.Map, roomName)
					}

				} else {
					gameServer.Log.Informational("creating room")
					confirm.Map[roomName] = make([]string, 2)
					confirm.Map[roomName][0] = so.Id()
				}
			}
		}

		aSendFunc := func(so socketio.Socket, roomName string) func(msg string) {

			gameLog.Lock()
			defer gameLog.Unlock()
			_, ok := gameLog.Map[roomName]
			if !ok {
				gameLog.Map[roomName] = make(map[string]string)
				gameLogTrue.Lock()
				defer gameLogTrue.Unlock()
				gameLogTrue.Map[roomName] = make([]string, 0)
			}

			return func(msg string) {
				gameServer.Log.Informational("%s(%s) said %s", so.Id(), so.Request().RemoteAddr, msg)
				gameServer.BroadcastTo(roomName, "info", "A selected ,pending B")
				so.BroadcastTo(roomName, "asend")
			}
		}

		bSendFunc := func(so socketio.Socket, roomName string) func(msg string) {
			return func(msg string) {
				result := strings.Split(msg, ",")
				gameLog.Map[roomName][result[0]] = result[1]
				if result[1] == "1" {
					gameLogTrue.Lock()
					defer gameLogTrue.Unlock()
					gameLogTrue.Map[roomName] = append(gameLogTrue.Map[roomName], result[0])
				}
				gameServer.Log.Informational("b send result to a %s", msg)
				gameServer.BroadcastTo(roomName, "info", "B selected ,pending A")
				so.BroadcastTo(roomName, "penda")
			}
		}
		gameOverFunc := func(so socketio.Socket, roomName string) func(msg string) {
			return func(msg string) {
				gameLogTrue.Lock()
				defer gameLogTrue.Unlock()

				jsonByte, _ := json.Marshal(gameLogTrue.Map[roomName])
				gameServer.BroadcastTo(roomName, "gameover", string(jsonByte))
				gameLogTrue.Map[roomName] = make([]string, 0)

			}
		}

		gameServer.RegisterEvent("gameover", gameOverFunc)
		gameServer.RegisterEvent("chatmsg", chatFunc)
		gameServer.RegisterEvent("confirm", confirmFunc)
		gameServer.RegisterEvent("asend", aSendFunc)
		gameServer.RegisterEvent("bsend", bSendFunc)
		gameServer.Start()
	}()
	beego.Run()
}
