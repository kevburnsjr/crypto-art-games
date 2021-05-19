package controller

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

func NewRouter(cfg *config.Api, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	rUser, err := repo.NewUser(cfg.Repo.User)
	if err != nil {
		logger.Fatal(err)
	}

	rFrame, err := repo.NewFrame(cfg.Repo.Frame)
	if err != nil {
		logger.Fatal(err)
	}

	rTileLock, err := repo.NewTileLock(cfg.Repo.TileLock)
	if err != nil {
		logger.Fatal(err)
	}

	rTileHistory, err := repo.NewTileHistory(cfg.Repo.TileHistory)
	if err != nil {
		logger.Fatal(err)
	}

	rUserFrameHistory, err := repo.NewUserFrameHistory(cfg.Repo.UserFrameHistory)
	if err != nil {
		logger.Fatal(err)
	}

	hub := sock.NewHub()
	go hub.Run()

	oauth := newOAuth(cfg, logger, rUser)

	socket := newSocket(logger, oauth, hub, rUser, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	debug := newDebug(cfg, logger, oauth, hub, rUser, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	router.Handle("/", index{oauth, cfg, logger, rUser})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/logout", newLogout(logger, oauth))
	router.Handle("/policy-accept", newPolicyAccept(logger, oauth, rUser))
	router.Handle("/privacy-policy", privacyPolicy{logger})
	router.Handle("/terms-of-service", termsOfService{logger})
	router.Handle("/oauth", oauth)
	router.Handle("/socket", socket)
	router.Handle("/debug", debug)
	router.NotFoundHandler = &static{"public"}

	return router
}
