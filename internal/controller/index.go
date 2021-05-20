package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
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

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	user, err := c.oauth.getUser(r, w)
	if err != ErrorTokenNotFound && check(err, w, c.log) {
		return
	}
	var userID uint16
	if user != nil {
		userID, _, err = c.repoUser.Find(user)
		if check(err, w, c.log) {
			return
		}
	}
	t, err := template.ParseFiles("./template/index.html")
	if check(err, w, c.log) {
		return
	}
	b := bytes.NewBuffer(nil)
	err = t.Execute(b, struct {
		UserID uint16
		User   *entity.User
	}{
		userID,
		user,
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
