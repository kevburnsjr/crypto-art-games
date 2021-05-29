package controller

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func newDebug(
	cfg *config.Api,
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
) *debug {
	return &debug{
		cfg:                  cfg,
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

type debug struct {
	cfg                  *config.Api
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

func (c *debug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	{
		username := "admin"
		password := c.cfg.Secret
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Who goes there?"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
	}
	user, err := c.oauth.getUser(r, w)
	if check(err, w, c.log) {
		return
	}

	var section = r.FormValue("section")
	var id = r.FormValue("id")
	var data string

	if r.Method == "POST" {
		switch section {
		case "series":
			id = r.FormValue("id")
			data = r.FormValue("data")
			series := entity.SeriesFromJson([]byte(data))
			if series == nil {
				println("invalid series")
				break
			}
			if id == "" {
				c.repoGame.InsertSeries(series)
			} else {
				c.repoGame.UpdateSeries(id, series)
			}
		}
	}
	if len(id) > 0 {
		switch section {
		case "series":
			s, err := c.repoGame.FindSeries(id)
			if err == nil {
				data = string(s.ToJson())
			}
		}
	}
	t := template.New("debug.html")
	t.Funcs(sprig.FuncMap())
	t.Funcs(map[string]interface{}{
		"pjson": func(v interface{}) template.JS {
			if vs, ok := v.(string); ok {
				m := map[string]interface{}{}
				json.Unmarshal([]byte(vs), &m)
				v = m
			}
			a, _ := json.MarshalIndent(v, "", "    ")
			return template.JS(a)
		},
		"hex": func(v interface{}) string {
			return hex.EncodeToString([]byte(v.(string)))
		},
	})
	_, err = t.ParseFiles("./template/debug.html")
	if check(err, w, c.log) {
		return
	}
	var boardId uint16 = 1
	b := bytes.NewBuffer(nil)
	err = t.Execute(b, struct {
		User         *entity.User
		Section      string
		BoardId      uint16
		RepoGame     repo.Game
		RepoUser     repo.User
		RepoBoard    repo.Board
		RepoReport   repo.Report
		RepoTileLock repo.TileLock
		Id           string
		Data         string
	}{
		user,
		section,
		boardId,
		c.repoGame,
		c.repoUser,
		c.repoBoard,
		c.repoReport,
		c.repoTileLock,
		id,
		data,
	})
	if check(err, w, c.log) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}
