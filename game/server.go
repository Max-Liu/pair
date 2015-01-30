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
	log      *logs.BeeLogger
	gameRoom map[string]int
}

type GameAdaptor struct {
	broadcast map[string]map[string]socketio.Socket
}

func (ga *GameAdaptor) Join(room string, socket socketio.Socket) error {
	sockets, ok := ga.broadcast[room]
	if !ok {
		sockets = make(map[string]socketio.Socket)
	}
	sockets[socket.Id()] = socket
	ga.broadcast[room] = sockets

	return nil
}

func (ga *GameAdaptor) GetServerNumber() (count int) {
	for k, _ := range ga.broadcast {
		count += len(ga.broadcast[k])
	}
	return count
}

func (ga *GameAdaptor) GetRoomNumber(room string) int {
	return len(ga.broadcast[room])
}

func (ga *GameAdaptor) Leave(room string, socket socketio.Socket) error {
	sockets, ok := ga.broadcast[room]
	if !ok {
		return nil
	}
	delete(sockets, socket.Id())
	if len(sockets) == 0 {
		delete(ga.broadcast, room)
		return nil
	}
	ga.broadcast[room] = sockets
	return nil
}
func (ga *GameAdaptor) Send(ignore socketio.Socket, room, message string, args ...interface{}) error {
	sockets := ga.broadcast[room]
	for id, s := range sockets {
		if ignore != nil && ignore.Id() == id {
			continue
		}
		s.Emit(message, args...)
	}
	return nil
}

var err error

func NewGameServer() *GameServer {
	gameServer := new(GameServer)
	gameServer.Server, err = socketio.NewServer(nil)
	gameServer.log = logs.NewLogger(10000)
	gameServer.log.SetLevel(log.Llongfile)
	gameServer.log.SetLogger("console", "")
	gameServer.gameRoom = make(map[string]int)
	gameAdaptor := new(GameAdaptor)
	gameAdaptor.broadcast = make(map[string]map[string]socketio.Socket)
	gameServer.SetAdaptor(gameAdaptor)
	return gameServer
}

func (s *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func (gameServer *GameServer) Start() {

	if err != nil {
		gameServer.log.Error("err:%s", err.Error())
	}

	gameServer.On("connection", func(so socketio.Socket) {
		input := context.NewInput(so.Request())
		chatRoom := input.Query("chat")
		gameServer.log.Informational("%s(%s) joined the room:%s", so.Id(), so.Request().RemoteAddr, chatRoom)

		so.Join(chatRoom)
		so.BroadcastTo(chatRoom, "joined", "your friend "+so.Id()+"joined the room.")
		peopleInRoom, ok := gameServer.gameRoom[chatRoom]

		if ok {
			gameServer.gameRoom[chatRoom] += 1
			if peopleInRoom >= 2 {
				so.Emit("info", "this room has fulled")
				so.BroadcastTo(chatRoom, "info", fmt.Sprintf("%s(%s) has quit the room", so.Id(), so.Request().RemoteAddr))
				gameServer.log.Informational("%s(%s) left the room:%s", so.Id(), so.Request().RemoteAddr, chatRoom)
				so.Leave(chatRoom)
			} else {
				gameServer.log.Informational("%d people in room:%s", gameServer.gameRoom[chatRoom], chatRoom)
			}
		} else {
			gameServer.gameRoom[chatRoom] = 1
			gameServer.log.Informational("%d people in room:%s", gameServer.gameRoom[chatRoom], chatRoom)
		}

		so.On("chat message", func(msg string) {
			gameServer.log.Informational("%s(%s) said %s", so.Id(), so.Request().RemoteAddr, msg)
			so.BroadcastTo(chatRoom, "chat message", msg)
		})

		so.On("disconnection", func() {
			gameServer.gameRoom[chatRoom] -= 1
			gameServer.log.Informational("%s(%s) disconnected", so.Id(), so.Request().RemoteAddr)
			gameServer.log.Informational("%d people in room:%s", gameServer.gameRoom[chatRoom], chatRoom)
			if gameServer.gameRoom[chatRoom] == 0 {
				gameServer.log.Informational("detoried the room %s", chatRoom)
				delete(gameServer.gameRoom, chatRoom)
			}

			so.BroadcastTo(chatRoom, "info", fmt.Sprintf("%s(%s) has quit the room", so.Id(), so.Request().RemoteAddr))
		})
	})

	gameServer.On("error", func(so socketio.Socket, err error) {
		gameServer.log.Error("err:%s", err.Error())
	})

	http.Handle("/socket.io/", gameServer)
	gameServer.log.Informational("Serving at localhost:5000")
	http.ListenAndServe(":5000", nil)
}
