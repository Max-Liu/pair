package main

import (
	"encoding/json"
	"fmt"
	"pair/game"
	_ "pair/routers"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	socketio "github.com/googollee/go-socket.io"
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
				peopleSlice, ok := confirm.Map[roomName]
				confirm.Unlock()
				if ok {
					if len(peopleSlice) == 2 {
						if peopleSlice[0] != so.Id() {
							peopleSlice[1] = so.Id()
							delete(confirm.Map, roomName)
							gameServer.Log.Informational("%s(%s) Game(%s) started.", so.Id(), so.Request().RemoteAddr, roomName)
							gameServer.BroadcastTo(roomName, "gamestart")
							go func() {
								go func() {
									var countSec int = 0
									for {
										gameServer.BroadcastTo(roomName, "info", "game started "+fmt.Sprintf("%d", countSec)+" Second")
										countSec += 5
										<-time.Tick(5 * time.Second)
										if countSec == 60 {
											gameServer.BroadcastTo(roomName, "info", "game started "+fmt.Sprintf("%d", countSec)+" Second")
											break
										}
									}
								}()
								<-time.Tick(60 * time.Second)
								gameLogTrue.Lock()
								defer gameLogTrue.Unlock()
								jsonByte, _ := json.Marshal(gameLogTrue.Map[roomName])
								gameServer.BroadcastTo(roomName, "gameover", string(jsonByte))
								gameServer.Log.Informational("Game Over(%s)", roomName)
								gameLogTrue.Map[roomName] = make([]string, 0)

							}()
						}
					}
					if len(peopleSlice) == 1 {
						delete(confirm.Map, roomName)
					}
					if len(peopleSlice) == 0 {
						delete(confirm.Map, roomName)
					}

				} else {
					confirm.Lock()
					gameServer.Log.Informational("creating room")
					confirm.Map[roomName] = make([]string, 2)
					confirm.Map[roomName][0] = so.Id()
					confirm.Unlock()
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
				so.BroadcastTo(roomName, "pendb", msg)
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
				gameServer.BroadcastTo(roomName, "penda", msg)
			}
		}
		//gameOverFunc := func(so socketio.Socket, roomName string) func(msg string) {
		//return func(msg string) {
		//gameLogTrue.Lock()
		//defer gameLogTrue.Unlock()

		//jsonByte, _ := json.Marshal(gameLogTrue.Map[roomName])
		//gameServer.BroadcastTo(roomName, "gameover", string(jsonByte))
		//gameLogTrue.Map[roomName] = make([]string, 0)

		//}
		//}

		//gameServer.RegisterEvent("gameover", gameOverFunc)
		gameServer.RegisterEvent("chatmsg", chatFunc)
		gameServer.RegisterEvent("confirm", confirmFunc)
		gameServer.RegisterEvent("asend", aSendFunc)
		gameServer.RegisterEvent("bsend", bSendFunc)
		gameServer.Start()
	}()
	beego.Run()
}
