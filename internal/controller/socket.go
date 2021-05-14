package controller

import (
	"context"
	"encoding/base64"
	"net/http"

	// "github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	// "github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func newSocket(logger *logrus.Logger, oauth *oauth, hub sock.Hub) *socket {
	return &socket{
		log:   logger,
		oauth: oauth,
		hub:   hub,
	}
}

type socket struct {
	log   *logrus.Logger
	oauth *oauth
	hub   sock.Hub
}

func (c socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := c.oauth.getSession(r)
	// lsn := r.FormValue("lsn")
	// timecode := r.FormValue("timecode")
	boardId := r.FormValue("boardId")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		c.log.Errorf("%w", err)
		return
	}

	conn := sock.CreateConnection([]string{boardId}, ws)
	/* Log replay
	if ts != "0" && ts != "" {
		logs, err := c.repoBoard.LogSince(boardId, lsn)
		if err == nil {
			log_objs := model.TeamLogParse(logs)
			for _, log := range log_objs {
				c.Write(websocket.TextMessage, socket.JsonMessage(team.Id, log))
			}
		}
	}
	*/

	ctx := context.WithValue(context.Background(), "session", session)
	ctx = context.WithValue(ctx, "boardId", boardId)

	c.hub.Register(conn)
	go conn.Writer()
	conn.Reader(c.hub, c.MsgHandler(ctx))
}

func (c socket) MsgHandler(ctx context.Context) sock.MessageHandler {
	return func(t int, msg []byte) {
		// Handle all operations and persist frames before broadcast
		if t == websocket.TextMessage {
			c.log.Debugf("Text: %#v", string(msg))
			c.hub.Broadcast(sock.TextMsgFromBytes(ctx.Value("boardId").(string), msg))
		} else if t == websocket.BinaryMessage {
			c.log.Debugf("Binary: %s", base64.StdEncoding.EncodeToString(msg))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(ctx.Value("boardId").(string), msg))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
	}
}
