package controller

import (
	"context"
	"encoding/json"
	"fmt"
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
	rUserBan repo.UserBan,
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
		repoUserBan:          rUserBan,
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
	repoUserBan          repo.UserBan
	repoFrame            repo.Frame
	repoTileLock         repo.TileLock
	repoTileHistory      repo.TileHistory
	repoUserFrameHistory repo.UserFrameHistory
}

func (c socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	user, err := c.oauth.getUser(r, w)
	// lsn := r.FormValue("lsn")
	userIdx := r.FormValue("userIdx")
	userBanIdx := r.FormValue("userBanIdx")
	boardId := r.FormValue("boardId")
	generation := r.FormValue("generation")
	timecode := r.FormValue("timecode")

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

	userIdxInt, _ := strconv.Atoi(userIdx)
	userBanIdxInt, _ := strconv.Atoi(userBanIdx)
	boardIdInt, _ := strconv.Atoi(boardId)
	generationInt, _ := strconv.Atoi(generation)
	timecodeInt, _ := strconv.Atoi(timecode)

	channels := []string{"global", "board-" + boardId}
	if user != nil && user.Policy {
		channels = append(channels, "user-"+strconv.Itoa(int(user.UserID)))
	}

	conn := sock.CreateConnection(channels, ws)

	// Sync new frames
	frames, err := c.repoFrame.Since(uint16(boardIdInt), uint16(generationInt), uint16(timecodeInt))
	for _, frame := range frames {
		conn.Write(sock.BinaryMsgFromBytes("board-"+boardId, frame.Data))
	}

	// Sync new users
	users, userIds, err := c.repoUser.Since(uint16(userIdxInt), uint16(generationInt))
	for i, user := range users {
		conn.Write(sock.TextMsgFromBytes("global", user.ToDto(userIds[i])))
	}

	// Sync new bans
	bans, err := c.repoUserBan.Since(uint16(userBanIdxInt))
	for i, user := range users {
		conn.Write(sock.TextMsgFromBytes("global", user.ToDto(userIds[i])))
	}

	conn.Write(sock.JsonMessage(boardId, map[string]interface{}{
		"type":       "sync-complete",
		"timecode":   timecodeInt + len(frames),
		"userIdx":    userIdxInt + len(users),
		"userBanIdx": userBanIdxInt + len(bans),
	}))

	ctx := context.WithValue(context.Background(), "user", user)
	ctx = context.WithValue(ctx, "userIdx", userIdx)
	ctx = context.WithValue(ctx, "boardId", "board-"+boardId)

	c.hub.Register(conn)
	go conn.Writer()
	conn.Reader(c.hub, c.MsgHandler(ctx))
}

func (c socket) MsgHandler(ctx context.Context) sock.MessageHandler {
	return func(t int, msg []byte) error {
		var userID uint16
		var user *entity.User
		var err error
		var found bool
		if u, ok := ctx.Value("user").(*entity.User); ok && u != nil {
			userID, found, err = c.repoUser.Find(u)
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("User authentication required")
			}
			user = u
		} else {
			return fmt.Errorf("User authentication required")
		}
		if !user.Policy {
			return fmt.Errorf("Must accept terms of service before participating")
		}
		if user.Timeout.After(time.Now()) {
			return fmt.Errorf("User banned")
		}

		boardId := ctx.Value("boardId").(string)

		// Handle all operations and persist frames before broadcast
		if t == websocket.TextMessage {
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
					return err
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardId, map[string]interface{}{
					"type":   "tile-locked",
					"tileID": tileID,
					"userID": userID,
				}))
			case "tile-lock-release":
				var tileID = uint16(m["tileID"].(float64))
				if err := c.repoTileLock.Release(userID, tileID, time.Now()); err != nil {
					return err
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardId, map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": userID,
				}))
			case "report-create":
				// Insert report
			case "report-clear":
				// Insert report
			case "frame-undo":
				// Insert frame undo
				// Mark frame hidden
				// Broadcast frame undo
			case "frame-redo":
				// Insert frame redo
				// Mark frame unhidden
				// Broadcast frame update
			case "user-ban":
				// Authorize user
				// Ban user on twitch
				// Insert ban
				// Mark user timed out
				// Mark relevant frames hidden
				// Broadcast user ban
			case "board-open":
				// unregister connection from previous board channel(s?)
				// register connection on new board channel(s?)
			}
			c.hub.Broadcast(sock.TextMsgFromBytes(boardId, msg))
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
			c.hub.Broadcast(sock.JsonMessagePure(boardId, map[string]interface{}{
				"type":   "tile-lock-release",
				"tileID": frame.TileID(),
				"userID": userID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(boardId, frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return nil
	}
}
