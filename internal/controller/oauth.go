package controller

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"github.com/nicklaw5/helix"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
)

type oauthError string

func (e oauthError) Error() string {
	return string(e)
}

const (
	stateCallbackKey  = "oauth-state-callback"
	oauthSessionName  = "oauth-oidc-session"
	oauthTokenKey     = "oauth-token"
	twitchUserDataKey = "twitch-user"

	ErrorTokenNotFound = oauthError("Token not found")
)

func newOAuth(cfg *config.Api, logger *logrus.Logger) *oauth {
	gob.Register(&oauth2.Token{})
	provider, err := oidc.NewProvider(context.Background(), cfg.Twitch.OidcIssuer)
	if err != nil {
		logger.Fatal(err)
	}
	return &oauth{
		log:          logger,
		cookieStore:  sessions.NewCookieStore([]byte(cfg.Secret)),
		oidcVerifier: provider.Verifier(&oidc.Config{ClientID: cfg.Twitch.ClientID}),
		oauth2Config: &oauth2.Config{
			ClientID:     cfg.Twitch.ClientID,
			ClientSecret: cfg.Twitch.ClientSecret,
			Scopes:       []string{oidc.ScopeOpenID, "user:read:email"},
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.Twitch.OAuthRedirect,
		},
	}
}

type oauth struct {
	log            *logrus.Logger
	cookieStore    *sessions.CookieStore
	oidcVerifier   *oidc.IDTokenVerifier
	oauth2Config   *oauth2.Config
}

func (c oauth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := c.cookieStore.Get(r, oauthSessionName)
	if err != nil {
		c.log.Warnf("corrupted session %s -- generated new", err)
		err = nil
	}

	errDesc := r.FormValue("error_description")
	if len(errDesc) > 0 {
		session.Save(r, w)
		http.Error(w, errDesc, 500)
		return
	}

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = fmt.Errorf("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'", state, stateChallenge[0])
	}

	if err != nil {
		session.Save(r, w)
		c.log.Warn(err.Error())
		http.Redirect(w, r, "/", 302)
		return
	}

	token, err := c.oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		session.Save(r, w)
		c.log.Error(err.Error())
		return
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = token

	c.log.Debugf("Access token: %#v", token)

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		session.Save(r, w)
		c.log.Warnf("can't extract id token from access token")
		http.Error(w, "Couldn't verify your confirmation, please try again.", 400)
		return
	}

	idToken, err := c.oidcVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		session.Save(r, w)
		http.Error(w, "Couldn't verify your confirmation, please try again.", 400)
		return
	}

	var claims struct {
		Iss   string `json:"iss"`
		Sub   string `json:"sub"`
		Aud   string `json:"aud"`
		Exp   int32  `json:"exp"`
		Iat   int32  `json:"iat"`
		Nonce string `json:"nonce"`
		Email string `json:"email"`
	}

	if err := idToken.Claims(&claims); err != nil {
		session.Save(r, w)
		http.Error(w, "Couldn't verify your confirmation, please try again.", 400)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
	return
}

func (c oauth) getSession(r *http.Request) (*sessions.Session, error) {
	return c.cookieStore.Get(r, oauthSessionName)
}

func (c oauth) getToken(r *http.Request) (*oauth2.Token, error) {
	session, err := c.cookieStore.Get(r, oauthSessionName)
	if err != nil {
		return nil, err
	}
	token, ok := session.Values[oauthTokenKey]
	if !ok {
		return nil, ErrorTokenNotFound
	}
	return token.(*oauth2.Token), nil
}

func (c oauth) getUser(r *http.Request, w http.ResponseWriter) (*entity.User, error) {
	session, err := c.cookieStore.Get(r, oauthSessionName)
	if err != nil {
		return nil, err
	}

	if userData, ok := session.Values[twitchUserDataKey]; ok {
		if u := entity.UserFromJson(userData.([]byte)); u != nil {
			return u, nil
		}
	}
	token, ok := session.Values[oauthTokenKey]
	if !ok {
		return nil, ErrorTokenNotFound
	}
	if err == nil && token != nil {
		client, err := helix.NewClient(&helix.Options{
			ClientID:        c.oauth2Config.ClientID,
			UserAccessToken: token.(*oauth2.Token).AccessToken,
		})
		if err != nil {
			return nil, err
		}
		resp, err := client.GetUsers(&helix.UsersParams{})
		if err != nil {
			return nil, err
		}
		if len(resp.Data.Users) > 0 {
			user := entity.User(resp.Data.Users[0])
			session.Values[twitchUserDataKey] = user.ToJson()
			session.Save(r, w)
			return &user, nil
		}
	}
	return nil, nil
}
