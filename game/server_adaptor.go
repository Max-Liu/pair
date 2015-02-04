package game

import "github.com/googollee/go-socket.io"

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
