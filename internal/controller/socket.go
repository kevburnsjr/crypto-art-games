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
) *socket {
	return &socket{
		log:          logger,
		oauth:        oauth,
		hub:          hub,
		repoGame:     rGame,
		repoUser:     rUser,
		repoLove:     rLove,
		repoBoard:    rBoard,
		repoFault:    rFault,
		repoReport:   rReport,
		repoUserBan:  rUserBan,
		repoTileLock: rTileLock,
	}
}

type socket struct {
	log          *logrus.Logger
	oauth        *oauth
	hub          sock.Hub
	repoGame     repo.Game
	repoUser     repo.User
	repoLove     repo.Love
	repoBoard    repo.Board
	repoFault    repo.Fault
	repoReport   repo.Report
	repoUserBan  repo.UserBan
	repoTileLock repo.TileLock
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

	channels := []string{"global", "bans"}
	if user != nil && user.Policy {
		channels = append(channels, "user-"+strconv.Itoa(int(user.UserID)))
		if user.Mod {
			channels = append(channels, "reports")
		}
	}

	conn := sock.CreateConnection(channels, ws)

	// Client thinks it's authed but user doesn't exist. Destroy session.
	if user != nil && user.Policy && !found {
		c.log.Errorf("User not found")
		conn.Write(sock.JsonMessage("", map[string]interface{}{
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

	banIdxInt, _ := strconv.Atoi(r.FormValue("banIdx"))
	userIdxInt, _ := strconv.Atoi(r.FormValue("userIdx"))
	var (
		reportIdx = 0
		banIdx    = uint32(banIdxInt)
		userIdx   = uint32(userIdxInt)
	)

	// Sync new reports for mods
	var reports []*entity.Report
	if user != nil && user.Mod {
		reports, err = c.repoReport.All()
	}
	for _, r := range reports {
		conn.Write(sock.TextMsgFromBytes("", r.ToDto()))
	}

	// Sync new user bans
	userBans, err := c.repoUserBan.Since(banIdx)
	for _, b := range userBans {
		conn.Write(sock.TextMsgFromBytes("", b.ToDto()))
		banIdx = b.ID
	}

	// Sync new users
	users, userIds, err := c.repoUser.Since(userIdx)
	if err != nil {
		c.log.Errorf("%v", err)
		http.Error(w, "Unable to retrieve new users", 500)
		return
	}
	for i, user := range users {
		conn.Write(sock.TextMsgFromBytes("", user.ToDto(userIds[i])))
		userIdx = user.UserID
	}

	conn.Write(sock.JsonMessage("", map[string]interface{}{
		"type":      "init",
		"v":         fmt.Sprintf("%016x", v),
		"user":      user,
		"userIdx":   userIdx,
		"banIdx":    banIdx,
		"reportIdx": reportIdx,
		"series":    series,
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

func (c socket) authMod(user *entity.User) (err error) {
	if err = c.auth(user); err != nil {
		return
	}
	if !user.Mod {
		err = fmt.Errorf("Unauthorized")
		return
	}
	return nil
}

func (c socket) MsgHandler(user *entity.User, conn sock.Connection) sock.MessageHandler {
	var board *entity.Board
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
				if err = c.repoTileLock.Acquire(user.UserID, boardId, tileID, time.Now()); err != nil {
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
				if err = c.repoTileLock.Release(user.UserID, boardId, tileID, time.Now()); err != nil {
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
			case "love":
				if err = c.auth(user); err != nil {
					return
				}
				var timecode = uint32(m["timecode"].(float64))
				var f *entity.Frame
				f, err = c.repoBoard.Find(boardId, timecode)
				if err != nil {
					return
				}
				if err = c.repoLove.Insert(boardId, timecode, user.UserID, time.Now()); err != nil {
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
			case "report":
				if err = c.auth(user); err != nil {
					return
				}
				var timecode = uint32(m["timecode"].(float64))
				var reason = m["reason"].(string)
				var f *entity.Frame
				if f, err = c.repoBoard.Find(boardId, timecode); err != nil {
					return
				}
				var report = &entity.Report{
					TargetID: f.UserID(),
					BoardID:  boardId,
					Timecode: timecode,
					UserID:   user.UserID,
					Date:     uint32(time.Now().Unix()),
					Reason:   reason,
				}
				if err = c.repoReport.Insert(report); err != nil {
					return
				}
				res = sock.NewJsonRes(report.ToResDto())
				c.hub.Broadcast(sock.NewJsonRes(report.ToDto()).Raw("reports"))
				return
			case "report-clear":
				if err = c.authMod(user); err != nil {
					return
				}
				var targetID = uint32(m["targetID"].(float64))
				if err = c.repoReport.Clear(targetID); err != nil {
					return
				}
				c.hub.Broadcast(sock.JsonMessagePure("reports", map[string]interface{}{
					"type":     "report-clear",
					"targetID": targetID,
				}))
			case "user-ban":
				if err = c.authMod(user); err != nil {
					return
				}
				var (
					targetID        = uint32(m["targetID"].(float64))
					since           = uint32(m["since"].(float64))
					timeout         = m["duration"].(string)
					reason          = ""    // m["reason"].(string)
					ban             = false // m["ban"].(bool)
					timeoutDuration time.Duration
				)
				if timeoutDuration, err = time.ParseDuration(timeout); err != nil && len(timeout) > 0 {
					return
				}
				var target *entity.User
				if target, err = c.repoUser.FindByUserID(targetID); err != nil {
					return
				}
				var userBan = entity.UserBan{
					ModID:    user.UserID,
					TargetID: targetID,
					Reason:   reason,
					Since:    since,
					Until:    uint32(time.Now().Add(timeoutDuration).Unix()),
					Ban:      ban,
				}
				if err = c.repoUserBan.Insert(&userBan); err != nil {
					return
				}
				if ban {
					target.Banned = true
				} else {
					target.Banned = false
					target.Timeout = time.Now().Add(timeoutDuration)
				}
				if err = c.repoUser.Update(target); err != nil {
					return
				}
				var n int
				var del int
				var allSeries entity.SeriesList
				if allSeries, err = c.repoGame.AllSeries(); err != nil {
					return
				}
				for _, s := range allSeries {
					for _, b := range s.Boards {
						if n, err = c.repoBoard.DeleteUserFramesAfter(b.ID, targetID, since); err != nil {
							return
						}
						del += n
					}
				}
				if err = c.repoReport.Clear(targetID); err != nil {
					return
				}
				c.hub.Broadcast(sock.NewJsonRes(userBan.ToDto()).Raw("bans"))
				c.hub.Broadcast(sock.JsonMessagePure("reports", map[string]interface{}{
					"type":     "report-clear",
					"targetID": targetID,
				}))
			case "err-storage":
				if err = c.auth(user); err != nil {
					return
				}
				err = c.repoFault.Insert("storage", uint32(m["userID"].(float64)), m["userAgent"].(string), time.Now())
				if err != nil {
					return
				}
			case "board-init":
				var (
					id       = uint16(m["boardId"].(float64))
					timecode = uint32(m["timecode"].(float64))
				)
				boardId = id
				boardChannel = fmt.Sprintf("board-%04x", boardId)
				board, err = c.repoGame.FindActiveBoard(boardId)
				if err != nil {
					return
				}
				channels := conn.Channels()
				for i, c := range channels {
					if strings.HasPrefix(c, "board-") {
						channels = append(channels[:i], channels[i+1:]...)
					}
				}
				channels = append(channels, boardChannel)
				c.hub.Update(conn, channels)
				// Sync new frames
				frames, err2 := c.repoBoard.Since(boardId, timecode)
				if err2 != nil {
					err = err2
					return
				}
				for _, frame := range frames {
					conn.Write(sock.BinaryMsgFromBytes(boardChannel, frame.Data))
					timecode = frame.Timestamp()
				}
				bucket := user.GetBucket(boardId)
				bucket.AdjustLevel(time.Now())
				conn.Write(sock.JsonMessage(boardChannel, map[string]interface{}{
					"type":     "board-init-complete",
					"timecode": timecode,
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
			if err = c.repoTileLock.Release(user.UserID, boardId, uint16(frame.TileID()), time.Now()); err != nil {
				// User does not have lock
				return
			}
			frame.SetTimestamp(uint32(time.Now().Unix()) - board.Created)
			if err = c.repoBoard.Insert(boardId, frame); err != nil {
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
