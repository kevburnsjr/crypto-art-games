package controller

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func newSocket(
	logger *logrus.Logger,
	oauth *oauth,
	hub sock.Hub,
	rUser repo.User,
	rFrame repo.Frame,
	rTileLock repo.TileLock,
	rTileHistory repo.TileHistory,
	rUserFrameHistory repo.UserFrameHistory,
) *socket {
	return &socket{
		log:                  logger,
		oauth:                oauth,
		hub:                  hub,
		repoUser:             rUser,
		repoFrame:            rFrame,
		repoTileLock:         rTileLock,
		repoTileHistory:      rTileHistory,
		repoUserFrameHistory: rUserFrameHistory,
	}
}

type socket struct {
	log                  *logrus.Logger
	oauth                *oauth
	hub                  sock.Hub
	repoUser             repo.User
	repoFrame            repo.Frame
	repoTileLock         repo.TileLock
	repoTileHistory      repo.TileHistory
	repoUserFrameHistory repo.UserFrameHistory
}

func (c socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := c.oauth.getUser(r, w)
	// lsn := r.FormValue("lsn")
	timecode := r.FormValue("timecode")
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

	tc, _ := strconv.Atoi(timecode)

	conn := sock.CreateConnection([]string{boardId}, ws)
	frames, err := c.repoFrame.Since(uint16(tc))
	for _, frame := range frames {
		time.Sleep(16 * time.Millisecond)
		conn.Write(sock.BinaryMsgFromBytes(boardId, frame.Data))
	}
	conn.Write(sock.JsonMessage(boardId, map[string]interface{}{
		"type":     "timecode-update",
		"timecode": tc + len(frames),
	}))

	ctx := context.WithValue(context.Background(), "user", user)
	ctx = context.WithValue(ctx, "boardId", boardId)

	c.hub.Register(conn)
	go conn.Writer()
	conn.Reader(c.hub, c.MsgHandler(ctx))
}

func (c socket) MsgHandler(ctx context.Context) sock.MessageHandler {
	return func(t int, msg []byte) error {
		var userID uint16
		var err error
		if u, ok := ctx.Value("user").(*entity.User); ok && u != nil {
			userID, err = c.repoUser.FindOrInsert(u)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
		// Handle all operations and persist frames before broadcast
		if t == websocket.TextMessage {
			c.log.Debugf("Text: %s", string(msg))
			var m = map[string]interface{}{}
			err := json.Unmarshal(msg, &m)
			if err != nil {
				return err
			}
			if _, ok := m["type"]; !ok {
				return err
			}
			switch m["type"].(string) {
			case "tile-lock":
				var tileID = uint16(m["tileID"].(float64))
				if err := c.repoTileLock.Acquire(userID, tileID, time.Now()); err != nil {
					// User does not have lock
					return err
				}
				c.hub.Broadcast(sock.JsonMessagePure(ctx.Value("boardId").(string), map[string]interface{}{
					"type":   "tile-locked",
					"tileID": tileID,
					"userID": userID,
				}))
			case "tile-lock-release":
				var tileID = uint16(m["tileID"].(float64))
				if err := c.repoTileLock.Release(userID, tileID, time.Now()); err != nil {
					// User does not have lock
					return err
				}
				c.hub.Broadcast(sock.JsonMessagePure(ctx.Value("boardId").(string), map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": userID,
				}))
			case "frame-undo":
				// Mark frame hidden
				// Broadcast frame update
			case "frame-redo":
				// Mark frame unhidden
				// Broadcast frame update
			}
			c.hub.Broadcast(sock.TextMsgFromBytes(ctx.Value("boardId").(string), msg))
		} else if t == websocket.BinaryMessage {
			frame := &entity.Frame{
				Data: msg,
			}
			frame.SetUserID(userID)
			if err = c.repoTileLock.Release(userID, frame.TileID(), time.Now()); err != nil {
				// User does not have lock
				return err
			}
			_, err = c.repoFrame.Insert(frame)
			if err != nil {
				return err
			}
			err = c.repoTileHistory.Insert(frame)
			if err != nil {
				return err
			}
			err = c.repoUserFrameHistory.Insert(frame)
			if err != nil {
				return err
			}
			c.hub.Broadcast(sock.JsonMessagePure(ctx.Value("boardId").(string), map[string]interface{}{
				"type":   "tile-lock-release",
				"tileID": frame.TileID(),
				"userID": userID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(ctx.Value("boardId").(string), frame.Data))
			c.log.Debugf("Binary: %s", hex.EncodeToString(frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return nil
	}
}
