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
	rBoard repo.Board,
	rLove repo.Love,
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
		repoBoard:            rBoard,
		repoLove:             rLove,
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
	repoBoard            repo.Board
	repoLove             repo.Love
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

	var boardIdUint16 = uint16(boardIdInt)
	var boardChannel = fmt.Sprintf("board-%04x", boardIdUint16)

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

	channels := []string{"global", boardChannel}
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

	// Sync new frames
	frames, err := c.repoBoard.Since(uint16(boardIdInt), uint16(generationInt), uint16(timecodeInt))
	var first *uint16
	if len(frames) > 0 {
		a := frames[0].Timecode()
		first = &a
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

	if user != nil && user.Policy {
		user.GetBucket(uint16(boardIdInt))
	}

	conn.Write(sock.JsonMessage("", map[string]interface{}{
		"type":     "init",
		"v":        fmt.Sprintf("%016x", v),
		"user":     user,
		"timecode": first,
		"series":   series,
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
	conn.Reader(c.hub, c.MsgHandler(user, boardIdUint16, boardChannel))
}

func (c socket) MsgHandler(user *entity.User, boardId uint16, boardChannel string) sock.MessageHandler {
	return func(t int, msg []byte) (res *sock.Msg, err error) {
		if user == nil {
			err = fmt.Errorf("User authentication required")
			return
		}
		var userID = user.UserID
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
				ftid, ok := m["tileID"].(float64)
				if !ok {
					err = fmt.Errorf("Malformed Tile ID %v", m["tileID"])
					return
				}
				var tileID = uint16(ftid)
				if err = c.repoUser.Consume(user, boardId); err != nil {
					return
				}
				if err = c.repoTileLock.Acquire(userID, tileID, time.Now()); err != nil {
					c.repoUser.Credit(user, boardId)
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-locked",
					"tileID": tileID,
					"userID": userID,
					"bucket": user.Buckets[boardId],
				}))
			case "tile-lock-release":
				var tileID = uint16(m["tileID"].(float64))
				if err = c.repoUser.Credit(user, boardId); err != nil {
					return
				}
				if err = c.repoTileLock.Release(userID, tileID, time.Now()); err != nil {
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": userID,
					"bucket": user.Buckets[boardId],
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
				if _, err = c.repoBoard.Find(boardId, timecode); err != nil {
					return
				}
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
			case "love":
				var timecode = uint16(m["timecode"].(float64))
				var f *entity.Frame
				f, err = c.repoBoard.Find(boardId, timecode)
				if err != nil {
					return
				}
				if err = c.repoLove.Insert(userID, timecode, time.Now()); err != nil {
					return
				}
				res = sock.NewJsonRes(map[string]interface{}{
					"type":     "love",
					"timecode": timecode,
					"userID":   userID,
				})
				c.hub.Broadcast(res.Raw("user-" + f.UserIDHex()))
				c.hub.Broadcast(res.Raw(boardChannel))
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
			case "board-switch":
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
				"userID": userID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return
	}
}
