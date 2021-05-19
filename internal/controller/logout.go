package controller

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func newLogout(logger *logrus.Logger, oauth *oauth) *logout {
	return &logout{
		log:   logger,
		oauth: oauth,
	}
}

type logout struct {
	log   *logrus.Logger
	oauth *oauth
}

func (c logout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := c.oauth.getSession(r)
	if err == nil {
		session.Options.MaxAge = -1
		err = session.Save(r, w)
	}
	http.Redirect(w, r, "/", 302)
}
