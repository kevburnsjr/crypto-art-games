package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
)

type index struct {
	*oauth
	cfg *config.Api
	log *logrus.Logger
}

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := c.oauth.getUser(r, w)
	if check(err, w, c.log) {
		return
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
		0,
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
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return true
	}
	return false
}
