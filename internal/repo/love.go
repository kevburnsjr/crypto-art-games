package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Love interface {
	Insert(boardID, timecode, userID uint16, t time.Time) (err error)
	All() (loves []*entity.Love, err error)
	Sweep(t time.Time) (s int, n int, err error)
}

// NewLove returns a Love repo instance
func NewLove(cfg config.KeyValueStore) (r *love, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &love{
		db: db,
	}, nil
}

type love struct {
	db driver.DB
}

// Insert inserts a love
func (r *love) Insert(boardID, timecode, userID uint16, t time.Time) (err error) {
	idBytes := make([]byte, 6)
	binary.BigEndian.PutUint16(idBytes[0:2], boardID)
	binary.BigEndian.PutUint16(idBytes[2:4], timecode)
	binary.BigEndian.PutUint16(idBytes[4:6], userID)
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val[0:4], uint32(t.Unix()))
	_, err = r.db.Put(idBytes, "", val)
	return
}

// All fetches all loves
func (r *love) All() (loves []*entity.Love, err error) {
	var start = make([]byte, 2)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, val := range vals {
		if len(keys[i]) != 4 {
			continue
		}
		loves = append(loves, &entity.Love{
			BoardID:  binary.BigEndian.Uint16(keys[i][0:2]),
			Timecode: binary.BigEndian.Uint16(keys[i][2:4]),
			UserID:   binary.BigEndian.Uint16(keys[i][4:6]),
			Date:     time.Unix(int64(binary.BigEndian.Uint32(val[0:4])), 0),
		})
	}
	return
}

// Sweep deletes all loves older than a given timestamp returning number scanned and number deleted
func (r *love) Sweep(t time.Time) (s int, n int, err error) {
	keys, vals, err := r.db.GetRanged([]byte(nil), 0, false)
	if err != nil {
		return
	}
	for i, val := range vals {
		if len(keys[i]) != 4 {
			continue
		}
		if time.Unix(int64(binary.BigEndian.Uint32(val[0:4])), 0).Before(t) {
			if err = r.db.Delete(keys[i], ""); err != nil {
				return
			}
			n++
		}
		s++
	}
	return
}

// Close closes a database connection
func (r *love) Close() {
	r.db.Close()
}
