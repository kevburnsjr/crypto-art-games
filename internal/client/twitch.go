package client

import (
	"github.com/sirupsen/logrus"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
)

type Twitch interface {
}

type twitch struct {
	cfg *config.Twitch
	log *logrus.Logger
}

func NewTwitch(cfg *config.Twitch, logger *logrus.Logger) Twitch {
	return &twitch{
		cfg: cfg,
		log: logger,
	}
}

func (c *twitch) WithOAuth() Twitch {
	return &twitch{
		cfg: c.cfg,
		log: c.log,
	}
}
