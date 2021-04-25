package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/controller"
)

type Api struct {
	config *config.Api
	logger *logrus.Logger
	server *http.Server
}

func NewApi(cfg *config.Api) *Api {
	return &Api{
		config: cfg,
		logger: newLogger(cfg.Log.Level),
	}
}

func (app *Api) Start() {
	cfg := app.config.Api
	handler := controller.NewRouter(app.config, app.logger)

	app.server = &http.Server{
		Handler: handler,
		Addr:    ":" + cfg.Port,
	}
	go func() {
		var err error
		app.logger.Printf("Api Listening on port %s", cfg.Port)
		if cfg.Ssl.Enabled {
			err = app.server.ListenAndServeTLS(cfg.Ssl.Cert, cfg.Ssl.Key)
		} else {
			err = app.server.ListenAndServe()
		}
		if err != nil {
			app.logger.Fatal(err.Error())
		}
	}()
}

func (app *Api) Stop(timeout time.Duration) {
	app.logger.Printf("Stopping HTTP Listener")
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	app.server.Shutdown(ctx)
}
