package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	rGame repo.Game,
	rUser repo.User,
	rLove repo.Love,
	rBoard repo.Board,
	rFault repo.Fault,
	rReport repo.Report,
	rUserBan repo.UserBan,
	rTileLock repo.TileLock,
	rTileHistory repo.TileHistory,
	rUserFrameHistory repo.UserFrameHistory,
) *socket {
	return &socket{
		log:                  logger,
		oauth:                oauth,
		hub:                  hub,
		repoGame:             rGame,
		repoUser:             rUser,
		repoLove:             rLove,
		repoBoard:            rBoard,
		repoFault:            rFault,
		repoReport:           rReport,
		repoUserBan:          rUserBan,
		repoTileLock:         rTileLock,
		repoTileHistory:      rTileHistory,
		repoUserFrameHistory: rUserFrameHistory,
	}
}

type socket struct {
	log                  *logrus.Logger
	oauth                *oauth
	hub                  sock.Hub
	repoGame             repo.Game
	repoUser             repo.User
	repoLove             repo.Love
	repoBoard            repo.Board
	repoFault            repo.Fault
	repoReport           repo.Report
	repoUserBan          repo.UserBan
	repoTileLock         repo.TileLock
	repoTileHistory      repo.TileHistory
	repoUserFrameHistory repo.UserFrameHistory
}

func (c socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	user, err := c.oauth.getUser(r, w)
	if err != nil && err != ErrorTokenNotFound {
		c.log.Errorf("%v", err)
		http.Error(w, "Unable to read session", 400)
		session, err := c.oauth.getSession(r)
		session.Options.MaxAge = -1
		session.Save(r, w)
		if err != nil {
			c.log.Errorf("%v", err)
		}
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var found bool
	if user != nil {
		_, found, err = c.repoUser.Find(user)
		if err != nil {
			c.log.Errorf("%v", err)
			http.Error(w, "Error finding user", 400)
			return
		}
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		c.log.Errorf("%v", err)
		return
	}

	channels := []string{"global"}
	if user != nil && user.Policy {
		channels = append(channels, "user-"+strconv.Itoa(int(user.UserID)))
	}

	conn := sock.CreateConnection(channels, ws)

	// Client thinks it's authed but user doesn't exist. Destroy session.
	if user != nil && user.Policy && !found {
		c.log.Errorf("User not found")
		conn.Write(sock.JsonMessage("global", map[string]interface{}{
			"type": "logout",
		}))
		return
	}

	v, err := c.repoGame.Version()
	if err != nil {
		c.log.Errorf("%v", err)
		http.Error(w, "Unable to retrieve game version", 500)
		return
	}

	series, err := c.repoGame.ActiveSeries()
	if err != nil {
		c.log.Errorf("%v", err)
		http.Error(w, "Unable to retrieve collections", 500)
		return
	}

	conn.Write(sock.JsonMessage("", map[string]interface{}{
		"type":   "init",
		"v":      fmt.Sprintf("%016x", v),
		"user":   user,
		"series": series,
	}))

	c.hub.Register(conn)
	go conn.Writer()
	conn.Reader(c.hub, c.MsgHandler(user, conn))
}

func (c socket) auth(user *entity.User) (err error) {
	if user == nil {
		err = fmt.Errorf("User authentication required")
		return
	}
	if !user.Policy {
		err = fmt.Errorf("Must accept terms of service before participating")
		return
	}
	if user.Timeout.After(time.Now()) {
		err = fmt.Errorf("User timed out")
		return
	}
	if user.Banned {
		err = fmt.Errorf("User banned")
		return
	}
	return nil
}

func (c socket) MsgHandler(user *entity.User, conn sock.Connection) sock.MessageHandler {
	var boardId uint16
	var boardChannel string
	return func(t int, msg []byte) (res *sock.Msg, err error) {
		// Handle all operations and persist frames before broadcast
		if t == websocket.TextMessage {
			var m = map[string]interface{}{}
			err = json.Unmarshal(msg, &m)
			if err != nil {
				return
			}
			if _, ok := m["type"]; !ok {
				return
			}
			switch m["type"].(string) {
			case "tile-lock":
				if err = c.auth(user); err != nil {
					return
				}
				ftid, ok := m["tileID"].(float64)
				if !ok {
					err = fmt.Errorf("Malformed Tile ID %v", m["tileID"])
					return
				}
				var tileID = uint16(ftid)
				if err = c.repoUser.Consume(user, boardId); err != nil {
					return
				}
				if err = c.repoTileLock.Acquire(user.UserID, tileID, time.Now()); err != nil {
					c.repoUser.Credit(user, boardId)
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-locked",
					"tileID": tileID,
					"userID": user.UserID,
					"bucket": user.Buckets[boardId],
				}))
			case "tile-lock-release":
				if err = c.auth(user); err != nil {
					return
				}
				var tileID = uint16(m["tileID"].(float64))
				if err = c.repoUser.Credit(user, boardId); err != nil {
					return
				}
				if err = c.repoTileLock.Release(user.UserID, tileID, time.Now()); err != nil {
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": user.UserID,
					"bucket": user.Buckets[boardId],
				}))
			case "frame-undo":
				if err = c.auth(user); err != nil {
					return
				}
				// Insert frame undo
				// Mark frame hidden
				// Broadcast frame undo
			case "frame-redo":
				if err = c.auth(user); err != nil {
					return
				}
				// Insert frame redo
				// Mark frame unhidden
				// Broadcast frame update
			case "report":
				if err = c.auth(user); err != nil {
					return
				}
				var timecode = uint16(m["timecode"].(float64))
				var reason = m["reason"].(string)
				if _, err = c.repoBoard.Find(boardId, timecode); err != nil {
					return
				}
				if err = c.repoReport.Insert(user.UserID, timecode, reason, time.Now()); err != nil {
					return
				}
				res = sock.NewJsonRes(map[string]interface{}{
					"type":     "report",
					"timecode": timecode,
					"userID":   user.UserID,
					"reason":   reason,
				})
				c.hub.Broadcast(res.Raw("reports"))
				return
			case "love":
				if err = c.auth(user); err != nil {
					return
				}
				var timecode = uint16(m["timecode"].(float64))
				var f *entity.Frame
				f, err = c.repoBoard.Find(boardId, timecode)
				if err != nil {
					return
				}
				if err = c.repoLove.Insert(user.UserID, timecode, time.Now()); err != nil {
					return
				}
				res = sock.NewJsonRes(map[string]interface{}{
					"type":     "love",
					"timecode": timecode,
					"userID":   user.UserID,
				})
				c.hub.Broadcast(res.Raw("user-" + f.UserIDHex()))
				c.hub.Broadcast(res.Raw(boardChannel))
				return
			case "report-clear":
				if err = c.auth(user); err != nil {
					return
				}
				// Clear report
			case "err-storage":
				if err = c.auth(user); err != nil {
					return
				}
				err = c.repoFault.Insert("storage", uint16(m["userID"].(float64)), m["userAgent"].(string), time.Now())
				if err != nil {
					return
				}
			case "user-ban":
				if err = c.auth(user); err != nil {
					return
				}
				// Authorize user
				// Ban user on twitch
				// Insert ban
				// Mark user timed out
				// Mark relevant frames hidden
				// Broadcast user ban
			case "board-init":
				var (
					id         = uint16(m["boardId"].(float64))
					userIdx    = uint16(m["userIdx"].(float64))
					generation = uint16(m["generation"].(float64))
					timecode   = uint16(m["timecode"].(float64))
				)
				boardId = id
				boardChannel = fmt.Sprintf("board-%04x", boardId)
				channels := conn.Channels()
				for i, c := range channels {
					if strings.HasPrefix(c, "board-") {
						channels = append(channels[:i], channels[i+1:]...)
					}
				}
				channels = append(channels, boardChannel)
				c.hub.Update(conn, channels)
				// Sync new users
				users, userIds, err2 := c.repoUser.Since(userIdx, generation)
				if err2 != nil {
					err = err2
					return
				}
				for i, user := range users {
					conn.Write(sock.TextMsgFromBytes("global", user.ToDto(userIds[i])))
					userIdx = user.UserID
				}
				// Sync new frames
				frames, err2 := c.repoBoard.Since(boardId, generation, timecode)
				if err2 != nil {
					err = err2
					return
				}
				for _, frame := range frames {
					conn.Write(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
					timecode = frame.Timecode() + 1
				}
				bucket := user.GetBucket(boardId)
				bucket.AdjustLevel(time.Now())
				conn.Write(sock.JsonMessage(boardChannel, map[string]interface{}{
					"type":     "board-init-complete",
					"timecode": timecode,
					"userIdx":  userIdx,
					"bucket":   bucket,
				}))
			}
			// c.hub.Broadcast(sock.TextMsgFromBytes(boardChannel, msg))
		} else if t == websocket.BinaryMessage {
			if err = c.auth(user); err != nil {
				return
			}
			frame := &entity.Frame{
				Data: msg,
			}
			frame.SetUserID(user.UserID)
			if err = c.repoTileLock.Release(user.UserID, frame.TileID(), time.Now()); err != nil {
				// User does not have lock
				return
			}

			_, err = c.repoBoard.Insert(boardId, frame, time.Now())
			if err != nil {
				return
			}
			err = c.repoTileHistory.Insert(frame)
			if err != nil {
				return
			}
			err = c.repoUserFrameHistory.Insert(frame)
			if err != nil {
				return
			}
			c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
				"type":   "tile-lock-release",
				"tileID": frame.TileID(),
				"userID": user.UserID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return
	}
}
