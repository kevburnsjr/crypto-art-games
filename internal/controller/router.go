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

	oauth := newOAuth(cfg, logger)

	hub := sock.NewHub()
	go hub.Run()

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

	socket := newSocket(logger, oauth, hub, rUser, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	debug := newDebug(logger, oauth, hub, rUser, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	router.Handle("/", index{oauth, cfg, logger, rUser})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/oauth", oauth)
	router.Handle("/socket", socket)
	router.Handle("/debug", debug)
	router.NotFoundHandler = &static{"public"}

	return router
}
