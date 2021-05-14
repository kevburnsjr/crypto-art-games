package controller

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
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

	rFrameLock, err := repo.NewFrameLock(cfg.Repo.FrameLock)
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

	socket := newSocket(logger, oauth, hub, rUser, rFrame, rFrameLock, rTileHistory, rUserFrameHistory)

	router.Handle("/", index{oauth, cfg, logger})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/oauth", oauth)
	router.Handle("/socket", socket)
	router.NotFoundHandler = &static{"public"}

	return router
}
