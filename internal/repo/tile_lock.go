package repo

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

var tileLockTimeout = 10 * time.Minute

type TileLock interface {
	Acquire(userID, tileID uint16, t time.Time) (err error)
	Release(userID, tileID uint16, t time.Time) (err error)
}

// NewTileLock returns an TileLock repo instance
func NewTileLock(cfg config.KeyValueStore) (r *tileLock, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &tileLock{
		db: db,
	}, nil
}

type tileLock struct {
	db driver.DB
}

// Acquire inserts a tileLock
func (r *tileLock) Acquire(userID, tileID uint16, t time.Time) (err error) {
	var key = make([]byte, 2)
	binary.BigEndian.PutUint16(key, tileID)

	_, b, err := r.db.Get(key)
	if err == errors.RepoItemNotFound {
		return r.acquire(key, userID, t)
	} else if err != nil {
		return
	} else if len(b) == 6 {
		if binary.BigEndian.Uint16(b[0:2]) == userID {
			return r.acquire(key, userID, t)
		}
		if time.Unix(int64(binary.BigEndian.Uint32(b[2:6])), 0).Before(t) {
			return r.acquire(key, userID, t)
		}
	}
	return fmt.Errorf("Tile locked")
}

func (r *tileLock) acquire(key []byte, userID uint16, t time.Time) (err error) {
	var val = make([]byte, 6)
	binary.BigEndian.PutUint16(val[0:2], userID)
	binary.BigEndian.PutUint32(val[2:6], uint32(t.Add(tileLockTimeout).Unix()))
	_, err = r.db.Put(key, "", val)
	return
}

// Release deletes a tileLock
func (r *tileLock) Release(userID, tileID uint16, t time.Time) (err error) {
	var key = make([]byte, 2)
	binary.BigEndian.PutUint16(key, tileID)

	vers, b, err := r.db.Get(key)
	if err == errors.RepoItemNotFound {
		return fmt.Errorf("Tile not locked %04x", tileID)
	} else if err != nil {
		return
	} else {
		//
		// This is really obnoxious. Need a better solution.
		//
		if len(b) == 6 && time.Unix(int64(binary.BigEndian.Uint32(b[2:6])), 0).Before(t) {
			return fmt.Errorf("Tile lock expired")
		}
	}
	if binary.BigEndian.Uint16(b[0:2]) != userID {
		return fmt.Errorf("Tile locked by another user")
	}
	return r.db.Delete(key, vers)
}

// All returns all records from the table
func (r *tileLock) All() (all map[string]string, err error) {
	all = map[string]string{}
	var start = make([]byte, 2)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, v := range vals {
		all[string(keys[i])] = string(v)
	}
	return
}

// Close closes a database connection
func (r *tileLock) Close() {
	r.db.Close()
}
