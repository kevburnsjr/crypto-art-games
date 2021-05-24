package socket

import (
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	timeSyncPeriod = 60 * time.Second
	maxMessageSize = 512
)

type Hub interface {
	Run()
	Broadcast(wsmessage)
	Register(*connection)
	Unregister(*connection)
}

type MessageHandler func(int, []byte) (*Msg, error)

func NewHub() *hub {
	return &hub{
		connections: make(map[string]map[*connection]bool),
		broadcast:   make(chan wsmessage),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
	}
}

type hub struct {
	connections map[string]map[*connection]bool
	broadcast   chan wsmessage
	register    chan *connection
	unregister  chan *connection
}

func (h *hub) Run() {
	for {
		select {
		case conn := <-h.register:
			for _, id := range conn.channel_ids {
				if _, ok := h.connections[id]; !ok {
					h.connections[id] = make(map[*connection]bool)
				}
				h.connections[id][conn] = true
			}
		case conn := <-h.unregister:
			for _, id := range conn.channel_ids {
				delete(h.connections[id], conn)
			}
			close(conn.send)
		case msg := <-h.broadcast:
			if connections, ok := h.connections[msg.channel_id]; ok {
				for conn := range connections {
					select {
					case conn.send <- msg:
					default:
						close(conn.send)
						delete(connections, conn)
					}
				}
			}
		}
	}
}

func (h *hub) Broadcast(msg wsmessage) {
	h.broadcast <- msg
}

func (h *hub) Register(conn *connection) {
	h.register <- conn
}

func (h *hub) Unregister(conn *connection) {
	h.unregister <- conn
}
