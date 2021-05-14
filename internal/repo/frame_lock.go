package repo

import (
	"encoding/binary"
	"fmt"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type FrameLock interface {
	Acquire(userID, tileID uint16) (err error)
	Release(userID, tileID uint16) (err error)
}

// NewFrameLock returns an FrameLock repo instance
func NewFrameLock(cfg config.KeyValueStore) (r *frameLock, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &frameLock{
		db: db,
	}, nil
}

type frameLock struct {
	db driver.DB
}

// Acquire inserts a frameLock
func (r *frameLock) Acquire(userID, tileID uint16) (err error) {
	var key = make([]byte, 2)
	binary.BigEndian.PutUint16(key, tileID)

	_, _, err = r.db.Get(key)
	if err == errors.RepoItemNotFound {
		var val = make([]byte, 2)
		binary.BigEndian.PutUint16(val, userID)
		_, err = r.db.Put(key, "", val)
		return
	} else if err != nil {
		return
	}
	return fmt.Errorf("Frame locked")
}

// Release deletes a frameLock
func (r *frameLock) Release(userID, tileID uint16) (err error) {
	var key = make([]byte, 2)
	binary.BigEndian.PutUint16(key, tileID)

	vers, b, err := r.db.Get(key)
	if err == errors.RepoItemNotFound {
		return nil
	} else if err != nil {
		return
	}
	if binary.BigEndian.Uint16(b) != userID {
		return fmt.Errorf("Frame locked by another user")
	}
	return r.db.Delete(key, vers)
}

// Close closes a database connection
func (r *frameLock) Close() {
	r.db.Close()
}
