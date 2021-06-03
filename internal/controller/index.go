package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

var allJS = []string{
	"/js/lib/helpers.js",
	"/js/lib/timeago.min.js",
	"/js/lib/jsuri-1.1.1.js",
	"/js/lib/base64.js",
	"/js/lib/bitset.min.js",
	"/js/lib/nearestColor.js",
	"/js/lib/localforage.min.js",
	"/js/lib/polyfills.js",
	"/js/global.js",
	"/js/game.js",
	"/js/lib/object.js",
	"/js/lib/event.js",
	"/js/lib/socket.js",
	"/js/lib/dom.js",
	"/js/series.js",
	"/js/user.js",
	"/js/nav.js",
	"/js/board.js",
	"/js/palette.js",
	"/js/tile.js",
	"/js/frame.js",
}

type index struct {
	*oauth
	cfg      *config.Api
	log      *logrus.Logger
	hub      sock.Hub
	repoUser repo.User
}

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	if r.URL.Path == "/" {
		w.Header().Set("Location", "/pixel-compactor")
		w.WriteHeader(302)
		return
	}
	var mod bool
	user, _ := c.oauth.getUser(r, w)
	if user != nil {
		mod = user.Mod
	}
	var js = allJS
	if c.cfg.Minify {
		js = []string{"/js/min.js?v="+c.cfg.Hash}
	}
	if c.cfg.Test {
		js = append(js, "/js/test.js")
	}

	b := bytes.NewBuffer(nil)
	var indexTpl = template.Must(template.ParseFiles("./template/index.html"))
	err := indexTpl.Execute(b, struct {
		HOST string
		JS   []string
		Mod  bool
	}{
		c.cfg.Http.Host,
		js,
		mod,
	})
	if check(err, w, c.log) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}

func check(err error, w http.ResponseWriter, log *logrus.Logger) bool {
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), 500)
		return true
	}
	return false
}
