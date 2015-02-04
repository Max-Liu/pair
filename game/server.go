package game

import (
	"fmt"
	"log"
	"net/http"

	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	socketio "github.com/googollee/go-socket.io"
)

type GameServer struct {
	*socketio.Server
	Log      *logs.BeeLogger
	gameRoom map[string]int
	baseConfig
	event map[string]func(so socketio.Socket, roomName string) func(msg string)
}

type baseConfig struct {
	RoomMaxNumber int
}

type con func(so socketio.Socket)

var err error

func NewGameServer() *GameServer {
	gameServer := new(GameServer)
	gameServer.Server, err = socketio.NewServer(nil)
	gameServer.Log = logs.NewLogger(10000)
	gameServer.Log.SetLevel(log.Llongfile)
	gameServer.Log.SetLogger("console", "")
	gameServer.gameRoom = make(map[string]int)
	gameAdaptor := new(GameAdaptor)
	gameAdaptor.broadcast = make(map[string]map[string]socketio.Socket)
	gameServer.SetAdaptor(gameAdaptor)
	gameServer.RoomMaxNumber = 2
	gameServer.event = make(map[string]func(so socketio.Socket, roomName string) func(msg string))
	return gameServer
}

func (gameServer *GameServer) checkPeopleNumberInRoom(roomName string, so socketio.Socket) {

	peopleInRoom, ok := gameServer.gameRoom[roomName]
	if ok {
		gameServer.gameRoom[roomName] += 1
		if peopleInRoom >= gameServer.RoomMaxNumber {
			so.Emit("info", "this room has fulled")
			so.BroadcastTo(roomName, "info", fmt.Sprintf("%s(%s) has quit the room", so.Id(), so.Request().RemoteAddr))
			gameServer.Log.Informational("%s(%s) left the room:%s", so.Id(), so.Request().RemoteAddr, roomName)
			so.Leave(roomName)
		} else {
			gameServer.Log.Informational("%d people in room:%s", gameServer.gameRoom[roomName], roomName)
		}
	} else {
		gameServer.gameRoom[roomName] = 1
		gameServer.Log.Informational("%d people in room:%s", gameServer.gameRoom[roomName], roomName)
	}
}

func (gameServer *GameServer) handleEvent() func(socketio.Socket) {
	return func(so socketio.Socket) {
		input := context.NewInput(so.Request())
		roomName := input.Query("chat")
		gameServer.Log.Informational("%s(%s) joined the room:%s", so.Id(), so.Request().RemoteAddr, roomName)
		so.Join(roomName)
		so.BroadcastTo(roomName, "joined", "your friend "+so.Id()+"joined the room.")
		gameServer.checkPeopleNumberInRoom(roomName, so)

		so.On("disconnection", func() {
			gameServer.gameRoom[roomName] -= 1
			gameServer.Log.Informational("%s(%s) disconnected", so.Id(), so.Request().RemoteAddr)
			gameServer.Log.Informational("%d people in room:%s", gameServer.gameRoom[roomName], roomName)
			if gameServer.gameRoom[roomName] == 0 {
				gameServer.Log.Informational("detoried the room %s", roomName)
				delete(gameServer.gameRoom, roomName)
			}
			so.BroadcastTo(roomName, "info", fmt.Sprintf("%s(%s) has quit the room", so.Id(), so.Request().RemoteAddr))
		})
		if len(gameServer.event) > 0 {
			for eventName, event := range gameServer.event {
				so.On(eventName, event(so, roomName))
			}
		}
	}
}

func (gameServer *GameServer) RegisterEvent(eventName string, event func(so socketio.Socket, roomName string) func(msg string)) {
	gameServer.event[eventName] = event
}

func (s *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func (gameServer *GameServer) Start() {
	if err != nil {
		gameServer.Log.Error("err:%s", err.Error())
	}
	gameServer.On("connection", gameServer.handleEvent())
	gameServer.On("error", func(so socketio.Socket, err error) {
		gameServer.Log.Error("err:%s", err.Error())
	})
	http.Handle("/socket.io/", gameServer)
	gameServer.Log.Informational("Serving at localhost:5000")
	http.ListenAndServe(":5000", nil)
}
