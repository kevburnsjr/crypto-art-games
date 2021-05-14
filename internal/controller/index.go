package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
)

type index struct {
	*oauth
	cfg *config.Api
	log *logrus.Logger
}

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := c.oauth.getUser(r, w)
	if c.check(w, err) {
		return
	}

	t, err := template.ParseFiles("./template/index.html")
	if c.check(w, err) {
		return
	}

	b := bytes.NewBuffer(nil)
	err = t.Execute(b, user)
	if c.check(w, err) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}

func (c index) check(w http.ResponseWriter, err error) bool {
	if err != nil {
		c.log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return true
	}
	return false
}
