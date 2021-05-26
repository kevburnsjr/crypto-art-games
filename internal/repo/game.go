package repo

import (
	"encoding/binary"
	"math/rand"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Game interface {
	Version() (v uint64, err error)
}

// NewGame returns an Game repo instance
func NewGame(cfg config.KeyValueStore) (r *game, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &game{
		db: db,
	}, nil
}

type game struct {
	db driver.DB
}

// Version retrieves or sets the game version
func (r *game) Version() (v uint64, err error) {
	_, vBytes, err := r.db.Get([]byte("_v"))
	if err == errors.RepoItemNotFound {
		vBytes = make([]byte, 8)
		rand.Read(vBytes)
		_, err = r.db.Put([]byte("_v"), "", vBytes)
		err = nil
	} else if err != nil {
		return
	} else {
		v = binary.BigEndian.Uint64(vBytes)
	}

	return
}

// Close closes a database connection
func (r *game) Close() {
	r.db.Close()
}
