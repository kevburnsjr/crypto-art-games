package controller

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
)

func NewRouter(cfg *config.Api, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	oauth := newOAuth(cfg, logger)

	router.Handle("/", index{oauth, cfg, logger})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/oauth", oauth)
	router.NotFoundHandler = &static{"public"}

	return router
}
