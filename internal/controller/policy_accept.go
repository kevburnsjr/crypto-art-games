package controller

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func newPolicyAccept(
	logger *logrus.Logger,
	oauth *oauth,
	hub sock.Hub,
	rUser repo.User,
) *policyAccept {
	return &policyAccept{
		log:      logger,
		oauth:    oauth,
		hub:      hub,
		repoUser: rUser,
	}
}

type policyAccept struct {
	log      *logrus.Logger
	oauth    *oauth
	hub      sock.Hub
	repoUser repo.User
}

func (c policyAccept) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	agree := r.FormValue("agree")
	if len(agree) < 1 {
		http.Error(w, "User must agree to terms of service and privacy policy in order to participate.", 400)
	}
	session, err := c.oauth.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	user, err := c.oauth.getUser(r, w)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	user.Policy = true
	user.Newsletter = len(r.FormValue("newsletter")) > 0
	userID, inserted, err := c.repoUser.FindOrInsert(user)
	if inserted {
		c.hub.Broadcast(sock.TextMsgFromBytes("global", user.ToDto(userID)))
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	session.Values[twitchUserDataKey] = user.ToJson()
	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
}
