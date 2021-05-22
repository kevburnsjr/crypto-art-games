package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type UserFrameHistory interface {
	Insert(frame *entity.Frame) (err error)
}

// NewUserFrameHistory returns an UserFrameHistory repo instance
func NewUserFrameHistory(cfg config.KeyValueStore) (r *userFrameHistory, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &userFrameHistory{
		db: db,
	}, nil
}

type userFrameHistory struct {
	db driver.DB
}

// Insert inserts a frame into the userFrame history
func (r *userFrameHistory) Insert(frame *entity.Frame) (err error) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint16(key[0:2], frame.UserID())
	binary.BigEndian.PutUint16(key[2:4], frame.Timecode())
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val, uint32(time.Now().Unix()))
	_, err = r.db.Put(key, "", val)
	if err != nil {
		return
	}
	return
}

// Close closes a database connection
func (r *userFrameHistory) Close() {
	r.db.Close()
}
