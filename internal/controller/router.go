package controller

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func NewRouter(cfg *config.Api, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	oauth := newOAuth(cfg, logger)

	hub := sock.NewHub()
	go hub.Run()
	socket := newSocket(logger, oauth, hub)

	router.Handle("/", index{oauth, cfg, logger})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/oauth", oauth)
	router.Handle("/socket", socket)
	router.NotFoundHandler = &static{"public"}

	return router
}
