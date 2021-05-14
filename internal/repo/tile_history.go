package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type TileHistory interface {
	Insert(frame *entity.Frame) (err error)
}

// NewTileHistory returns an TileHistory repo instance
func NewTileHistory(cfg config.KeyValueStore) (r *tileHistory, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &tileHistory{
		db: db,
	}, nil
}

type tileHistory struct {
	db driver.DB
}

// Insert inserts a frame into the tile history
func (r *tileHistory) Insert(frame *entity.Frame) (err error) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint16(key[0:2], frame.TileID())
	binary.BigEndian.PutUint16(key[2:4], frame.Timecode)
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val, uint32(time.Now().Unix()))
	_, err = r.db.Put(key, "", val)
	if err != nil {
		return
	}
	return
}

// Close closes a database connection
func (r *tileHistory) Close() {
	r.db.Close()
}
