package socket

import (
	"time"

	"github.com/gorilla/websocket"
)

type Connection interface {
	Write(m wsmessage) error
	Channels() []string
}

func CreateConnection(ids []string, ws *websocket.Conn) *connection {
	return &connection{channel_ids: ids, ws: ws, send: make(chan wsmessage, 256)}
}

type connection struct {
	channel_ids []string
	ws          *websocket.Conn
	send        chan wsmessage
}

func (c *connection) Reader(hub Hub, handler MessageHandler) {
	defer func() {
		hub.Unregister(c)
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		t, b, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		res, err := handler(t, b)
		if err != nil {
			c.send <- JsonMessagePure("", map[string]interface{}{
				"type": "err",
				"msg":  err.Error(),
			})
		} else if res != nil {
			c.send <- wsmessage(*res)
		}
	}
}

func (c *connection) Write(m wsmessage) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(m.msgType, m.data)
}

func (c *connection) Channels() []string {
	return c.channel_ids
}

func (c *connection) hasChannel(ch1 string) bool {
	for _, ch2 := range c.channel_ids {
		if ch1 == ch2 {
			return true
		}
	}
	return false
}

func (c *connection) Writer() {
	pinger := time.NewTicker(pingPeriod)
	defer func() {
		pinger.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.Write(wsmessage{websocket.CloseMessage, "", nil})
				return
			}
			if err := c.Write(message); err != nil {
				return
			}
		case <-pinger.C:
			if err := c.Write(wsmessage{websocket.PingMessage, "", nil}); err != nil {
				return
			}
		}
	}
}
