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
	if err != nil && err != ErrorTokenNotFound {
		c.log.Errorf("%w", err)
		http.Error(w, "Unable to read session", 400)
		session, err := c.oauth.getSession(r)
		if err == nil {
			session.Options.MaxAge = -1
			session.Save(r, w)
		}
		return
	}

	userIdx := r.FormValue("userIdx")
	userBanIdx := r.FormValue("userBanIdx")
	boardId := r.FormValue("boardId")
	generation := r.FormValue("generation")
	timecode := r.FormValue("timecode")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	userIdxInt, _ := strconv.Atoi(userIdx)
	userBanIdxInt, _ := strconv.Atoi(userBanIdx)
	boardIdInt, _ := strconv.Atoi(boardId)
	generationInt, _ := strconv.Atoi(generation)
	timecodeInt, _ := strconv.Atoi(timecode)

	if user != nil {
		var found bool
		_, found, err = c.repoUser.Find(user)
		if err != nil {
			c.log.Errorf("%w", err)
			http.Error(w, "Error finding user", 400)
			return
		}
		if user.Policy && !found {
			session, err := c.oauth.getSession(r)
			if err == nil {
				session.Options.MaxAge = -1
				session.Save(r, w)
			}
			http.Error(w, "User not found", 400)
			return
		}
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		c.log.Errorf("%w", err)
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

	// Sync new frames
	frames, err := c.repoFrame.Since(uint16(boardIdInt), uint16(generationInt), uint16(timecodeInt))
	var first *uint16
	if len(frames) > 0 {
		a := frames[0].Timecode()
		first = &a
	}

	conn := sock.CreateConnection(channels, ws)

	conn.Write(sock.JsonMessage(boardId, map[string]interface{}{
		"type":     "init",
		"user":     user,
		"timecode": first,
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
	return func(t int, msg []byte) error {
		var userID = user.UserID
		var err error
		if user == nil {
			return fmt.Errorf("User authentication required")
		}
		if !user.Policy {
			return fmt.Errorf("Must accept terms of service before participating")
		}
		if user.Timeout.After(time.Now()) {
			return fmt.Errorf("User timed out")
		}
		if user.Banned {
			return fmt.Errorf("User banned")
		}

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
				if err = c.repoUser.Consume(user); err != nil {
					return err
				}
				if err := c.repoTileLock.Acquire(userID, tileID, time.Now()); err != nil {
					return err
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
					return err
				}
				if err := c.repoTileLock.Release(userID, tileID, time.Now()); err != nil {
					return err
				}
				c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
					"type":   "tile-lock-released",
					"tileID": tileID,
					"userID": userID,
					"bucket": user.Bucket,
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
				// Could avoid this by just breaking the socket and reconnecting.
			}
			c.hub.Broadcast(sock.TextMsgFromBytes(boardChannel, msg))
		} else if t == websocket.BinaryMessage {
			frame := &entity.Frame{
				Data: msg,
			}
			frame.SetUserID(userID)
			if err = c.repoTileLock.Release(userID, frame.TileID(), time.Now()); err != nil {
				// User does not have lock
				return err
			}
			_, err = c.repoFrame.Insert(frame, time.Now())
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
			c.hub.Broadcast(sock.JsonMessagePure(boardChannel, map[string]interface{}{
				"type":   "tile-lock-release",
				"tileID": frame.TileID(),
				"userID": userID,
			}))
			c.hub.Broadcast(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
		} else {
			c.log.Debugf("Uknown: %d, %s", t, string(msg))
		}
		return nil
	}
}
