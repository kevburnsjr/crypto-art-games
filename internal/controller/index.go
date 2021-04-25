package controller

import (
	"bytes"
	"net/http"
	"html/template"

	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
)

type index struct{
	*oauth
	cfg *config.Api
	log *logrus.Logger

}

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user helix.User
	token, err := c.oauth.getToken(r)
	if err == nil && token != nil {
		client, err := helix.NewClient(&helix.Options{
			ClientID:        c.cfg.Twitch.ClientID,
			UserAccessToken: token.AccessToken,
		})
		if c.check(w, err) {
			return
		}
		resp, err := client.GetUsers(&helix.UsersParams{})
		if c.check(w, err) {
			return
		}
		if len(resp.Data.Users) > 0 {
			user = resp.Data.Users[0]
		}
	}

	t, err := template.ParseFiles("./template/index.html")
	if c.check(w, err) {
		return
	}

	b := bytes.NewBuffer(nil)
	err = t.Execute(b, user)
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