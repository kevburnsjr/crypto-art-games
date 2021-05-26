package controller

import (
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
	rGame repo.Game,
	rUser repo.User,
	rReport repo.Report,
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
		repoGame:             rGame,
		repoUser:             rUser,
		repoReport:           rReport,
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
	repoGame             repo.Game
	repoUser             repo.User
	repoReport           repo.Report
	repoUserBan          repo.UserBan
	repoFrame            repo.Frame
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

	var (
		boardId    = r.FormValue("boardId")
		generation = r.FormValue("generation")
		// palette    = r.FormValue("palette")
		timecode   = r.FormValue("timecode")
		userBanIdx = r.FormValue("userBanIdx")
		userIdx    = r.FormValue("userIdx")
	)

	boardIdInt, _ := strconv.Atoi(boardId)
	generationInt, _ := strconv.Atoi(generation)
	timecodeInt, _ := strconv.Atoi(timecode)
	userBanIdxInt, _ := strconv.Atoi(userBanIdx)
	userIdxInt, _ := strconv.Atoi(userIdx)

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

	channels := []string{"global", "board-" + boardId}
	if user != nil && user.Policy {
		channels = append(channels, "user-"+strconv.Itoa(int(user.UserID)))
		if user.Bucket == nil {
			user.Bucket = entity.NewUserBucket()
		} else {
			user.Bucket.AdjustLevel()
		}
	}

	conn := sock.CreateConnection(channels, ws)

	if user != nil && user.Policy && !found {
		c.log.Errorf("User not found")
		conn.Write(sock.JsonMessage("global", map[string]interface{}{
			"type": "logout",
		}))
		return
	}

	// Sync new frames
	frames, err := c.repoFrame.Since(uint16(boardIdInt), uint16(generationInt), uint16(timecodeInt))
	var first *uint16
	if len(frames) > 0 {
		a := frames[0].Timecode()
		first = &a
	}

	v, err := c.repoGame.Version()
	if err != nil {
		c.log.Errorf("%v", err)
		http.Error(w, "Unable to retrieve game vesrion", 500)
		return
	}

	conn.Write(sock.JsonMessage(boardId, map[string]interface{}{
		"type":     "init",
		"user":     user,
		"timecode": first,
		"v":        fmt.Sprintf("%016x", v),
	}))

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
	for _, ban := range bans {
		conn.Write(sock.TextMsgFromBytes("global", ban.ToDto()))
	}

	conn.Write(sock.JsonMessage(boardId, map[string]interface{}{
		"type":       "sync-complete",
		"timecode":   timecodeInt + len(frames),
		"userIdx":    userIdxInt + len(users),
		"userBanIdx": userBanIdxInt + len(bans),
	}))

	c.hub.Register(conn)
	go conn.Writer()
	conn.Reader(c.hub, c.MsgHandler(user, "board-"+boardId))
}

func (c socket) MsgHandler(user *entity.User, boardChannel string) sock.MessageHandler {
	return func(t int, msg []byte) (res *sock.Msg, err error) {
		var userID = user.UserID
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
				var tileID = uint16(m["tileID"].(float64))
				if err = c.repoUser.Consume(user); err != nil {
					return
				}
				if err = c.repoTileLock.Acquire(userID, tileID, time.Now()); err != nil {
					c.repoUser.Credit(user)
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-locked",
					"tileID": tileID,
					"userID": userID,
					"bucket": user.Bucket,
				}))
			case "tile-lock-release":
				var tileID = uint16(m["tileID"].(float64))
				if err = c.repoUser.Credit(user); err != nil {
					return
				}
				if err = c.repoTileLock.Release(userID, tileID, time.Now()); err != nil {
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": userID,
					"bucket": user.Bucket,
				}))
			case "frame-undo":
				// Insert frame undo
				// Mark frame hidden
				// Broadcast frame undo
			case "frame-redo":
				// Insert frame redo
				// Mark frame unhidden
				// Broadcast frame update
			case "report":
				var timecode = uint16(m["timecode"].(float64))
				var reason = m["reason"].(string)
				if err = c.repoReport.Insert(userID, timecode, reason, time.Now()); err != nil {
					return
				}
				res = sock.NewJsonRes(map[string]interface{}{
					"type":     "report",
					"timecode": timecode,
					"userID":   userID,
					"reason":   reason,
				})
				c.hub.Broadcast(res.Raw("reports"))
				return
			case "report-clear":
				// Insert report
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
				// Could avoid this by just breaking the socket and reconnecting.
			}
			// c.hub.Broadcast(sock.TextMsgFromBytes(boardChannel, msg))
		} else if t == websocket.BinaryMessage {
			frame := &entity.Frame{
				Data: msg,
			}
			frame.SetUserID(userID)
			if err = c.repoTileLock.Release(userID, frame.TileID(), time.Now()); err != nil {
				// User does not have lock
				return
			}
			_, err = c.repoFrame.Insert(frame, time.Now())
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
				"userID": userID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return
	}
}
