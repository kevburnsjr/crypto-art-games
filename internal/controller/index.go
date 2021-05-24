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

type index struct {
	*oauth
	cfg      *config.Api
	log      *logrus.Logger
	hub      sock.Hub
	repoUser repo.User
}

var indexTpl = template.Must(template.ParseFiles("./template/index.html"))

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	b := bytes.NewBuffer(nil)
	err := indexTpl.Execute(b, struct {
		HOST string
	}{
		c.cfg.Api.Host,
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
