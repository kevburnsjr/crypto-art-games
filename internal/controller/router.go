package controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

var stdHeaders func(w http.ResponseWriter)

func NewRouter(cfg *config.Api, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	rGame, err := repo.NewGame(cfg.Repo.Game)
	if err != nil {
		logger.Fatal(err)
	}

	rUser, err := repo.NewUser(cfg.Repo.User)
	if err != nil {
		logger.Fatal(err)
	}

	rBoard, err := repo.NewBoard(cfg.Repo.Board)
	if err != nil {
		logger.Fatal(err)
	}

	rLove, err := repo.NewLove(cfg.Repo.Love)
	if err != nil {
		logger.Fatal(err)
	}

	rFault, err := repo.NewFault(cfg.Repo.Fault)
	if err != nil {
		logger.Fatal(err)
	}

	rReport, err := repo.NewReport(cfg.Repo.Report)
	if err != nil {
		logger.Fatal(err)
	}

	rUserBan, err := repo.NewUserBan(cfg.Repo.UserBan)
	if err != nil {
		logger.Fatal(err)
	}

	rTileLock, err := repo.NewTileLock(cfg.Repo.TileLock)
	if err != nil {
		logger.Fatal(err)
	}

	imgUrl := "https://static-cdn.jtvnw.net"
	wsUrl := "wss://" + cfg.Http.Host

	stdHeaders = func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: "+imgUrl+"; font-src 'self' data:; frame-src; script-src 'self'; style-src 'self'; connect-src 'self' "+wsUrl)
	}

	hub := sock.NewHub()
	go hub.Run()

	oauth := newOAuth(cfg, logger, rUser)

	socket := newSocket(logger, oauth, hub, rGame, rUser, rLove, rBoard, rFault, rReport, rUserBan, rTileLock)

	debug := newDebug(cfg, logger, oauth, hub, rGame, rUser, rLove, rBoard, rFault, rReport, rUserBan, rTileLock)

	router.Handle("/", index{})
	router.Handle("/pixel-compactor", index{oauth, cfg, logger, hub, rUser})
	router.Handle("/u/i/{id:[0-9]+}", newUserImage(rUser))
	router.Handle("/js/min.js", &staticMinJS{"public", cfg.Hash})
	router.Handle("/login", newLogin(logger, oauth))
	router.Handle("/logout", newLogout(logger, oauth))
	router.Handle("/policy-accept", newPolicyAccept(logger, oauth, hub, rUser))
	router.Handle("/privacy-policy", privacyPolicy{logger})
	router.Handle("/terms-of-service", termsOfService{logger})
	router.Handle("/oauth", oauth)
	router.Handle("/socket", socket)
	router.Handle("/debug", debug)
	router.NotFoundHandler = &static{"public"}

	return router
}
