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
	rUser repo.User,
	rFrame repo.Frame,
	rTileLock repo.TileLock,
	rTileHistory repo.TileHistory,
	rUserFrameHistory repo.UserFrameHistory,
) *debug {
	return &debug{
		cfg:                  cfg,
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

type debug struct {
	cfg                  *config.Api
	log                  *logrus.Logger
	oauth                *oauth
	hub                  sock.Hub
	repoUser             repo.User
	repoFrame            repo.Frame
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

	section := r.FormValue("section")
	user, err := c.oauth.getUser(r, w)
	if check(err, w, c.log) {
		return
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
	b := bytes.NewBuffer(nil)
	err = t.Execute(b, struct {
		User         *entity.User
		Section      string
		RepoUser     repo.User
		RepoFrame    repo.Frame
		RepoTileLock repo.TileLock
	}{
		user,
		section,
		c.repoUser,
		c.repoFrame,
		c.repoTileLock,
	})
	if check(err, w, c.log) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}
