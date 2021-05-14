package controller

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func newLogin(logger *logrus.Logger, oauth *oauth) *login {
	return &login{
		log:   logger,
		oauth: oauth,
	}
}

type login struct {
	log   *logrus.Logger
	oauth *oauth
}

func (c login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := c.oauth.getSession(r)

	var tokenBytes [255]byte
	if _, err = rand.Read(tokenBytes[:]); err != nil {
		c.log.Error("Couldn't generate a session - %s", err.Error())
		http.Error(w, "Couldn't generate a session", 500)
		return
	}
	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err = session.Save(r, w); err != nil {
		c.log.Error("Couldn't save session - %s", err.Error())
		http.Error(w, "Couldn't save session", 500)
		return
	}
	claims := oauth2.SetAuthURLParam("claims", `{"id_token":{"email":null}}`)
	http.Redirect(w, r, c.oauth.oauth2Config.AuthCodeURL(state, claims), 302)
}
