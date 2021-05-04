package controller

import (
	"encoding/base64"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"

	// "github.com/kevburnsjr/crypto-art-games/internal/repo"
	// "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func newSocket(logger *logrus.Logger, oauth *oauth) *socket {
	return &socket{
		log:   logger,
		oauth: oauth,
	}
}

type socket struct {
	log   *logrus.Logger
	oauth *oauth
}

func (c socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// session, err := c.oauth.getSession(r)
	// lsn := r.FormValue("lsn")
	// timecode := r.FormValue("timecode")
	// boardId := r.FormValue("boardId")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		http.Error(w, "Not a websocket handshake", 400)
		return
	}
	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				c.log.Errorln(err)
				return
			}
			c.log.Debugln("Receive:", base64.StdEncoding.EncodeToString(msg), op, err)
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				c.log.Errorln(err)
				return
			}
		}
	}()

	wsutil.WriteServerMessage(conn, ws.OpPing, []byte(nil))

	// socket.TeamHub.Register <- c
	// go c.Writer()
	// c.Reader(socket.TeamHub)
}
