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

	rUser, err := repo.NewUser(cfg.Repo.User)
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

	imgUrl := "https://static-cdn.jtvnw.net"
	wsUrl := "wss://" + cfg.Api.Host

	stdHeaders = func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: "+imgUrl+"; font-src 'self' data:; frame-src https://www.twitch.tv; script-src 'self'; style-src 'self'; connect-src 'self' "+wsUrl)
	}

	hub := sock.NewHub()
	go hub.Run()

	oauth := newOAuth(cfg, logger, rUser)

	socket := newSocket(logger, oauth, hub, rUser, rReport, rUserBan, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	debug := newDebug(cfg, logger, oauth, hub, rUser, rReport, rUserBan, rFrame, rTileLock, rTileHistory, rUserFrameHistory)

	router.Handle("/", index{})
	router.Handle("/1", index{oauth, cfg, logger, hub, rUser})
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
