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
	Acquire(userID uint32, boardID, tileID uint16, t time.Time) (err error)
	Release(userID uint32, boardID, tileID uint16, t time.Time) (err error)
}

// NewTileLock returns an TileLock repo instance
func NewTileLock(cfg config.KeyValueStore) (r *tileLock, err error) {
	var tileDB driver.DB
	var userDB driver.DB
	if cfg.LevelDB != nil {
		tileDBCfg := *cfg.LevelDB
		tileDBCfg.Path += "-tile"
		tileDB, err = driver.NewLevelDB(tileDBCfg)
		userDBCfg := *cfg.LevelDB
		userDBCfg.Path += "-user"
		userDB, err = driver.NewLevelDB(userDBCfg)
	}
	if err != nil || tileDB == nil {
		return
	}
	return &tileLock{
		tileDB: tileDB,
		userDB: userDB,
	}, nil
}

type tileLock struct {
	tileDB driver.DB
	userDB driver.DB
}

// Acquire inserts a tileLock
func (r *tileLock) Acquire(userID uint32, boardID, tileID uint16, t time.Time) (err error) {
	// Clear any existing locks belonging to this user
	var userKey = make([]byte, 4)
	binary.BigEndian.PutUint32(userKey[0:4], userID)
	_, b, err := r.userDB.Get(userKey)
	if err == errors.RepoItemNotFound {
		if err = r.tileDB.Delete(b, ""); err != nil {
			return
		}
	} else if err != nil {
		return
	}

	var key = make([]byte, 4)
	binary.BigEndian.PutUint16(key[0:2], boardID)
	binary.BigEndian.PutUint16(key[2:4], tileID)

	_, val, err := r.tileDB.Get(key)
	if err == errors.RepoItemNotFound {
		return r.acquire(key, userID, t)
	} else if err != nil {
		return
	} else if len(val) == 8 {
		if binary.BigEndian.Uint32(val[0:4]) == userID {
			return r.acquire(key, userID, t)
		}
		if time.Unix(int64(binary.BigEndian.Uint32(val[4:8])), 0).Before(t) {
			return r.acquire(key, userID, t)
		}
	}
	return fmt.Errorf("Tile locked")
}

func (r *tileLock) acquire(key []byte, userID uint32, t time.Time) (err error) {
	var val = make([]byte, 8)
	binary.BigEndian.PutUint32(val[0:4], userID)
	binary.BigEndian.PutUint32(val[4:8], uint32(t.Add(tileLockTimeout).Unix()))
	if _, err = r.tileDB.Put(key, "", val); err != nil {
		return
	}
	if _, err = r.userDB.Put(val[0:4], "", key); err != nil {
		return
	}
	return
}

// Release deletes a tileLock
func (r *tileLock) Release(userID uint32, boardID uint16, tileID uint16, t time.Time) (err error) {
	var key = make([]byte, 4)
	binary.BigEndian.PutUint16(key[0:2], boardID)
	binary.BigEndian.PutUint16(key[2:4], uint16(tileID))

	vers, val, err := r.tileDB.Get(key)
	if err == errors.RepoItemNotFound {
		return fmt.Errorf("Tile not locked %02x", tileID)
	} else if err != nil {
		return
	} else {
		//
		// This is really obnoxious. Need a better solution.
		//
		if len(val) == 8 && time.Unix(int64(binary.BigEndian.Uint32(val[4:8])), 0).Before(t) {
			return fmt.Errorf("Tile lock expired")
		}
	}
	if binary.BigEndian.Uint32(val[0:4]) != userID {
		return fmt.Errorf("Tile locked by another user")
	}
	if err = r.userDB.Delete(val[0:4], ""); err != nil {
		return
	}

	return r.tileDB.Delete(key, vers)
}

// All returns all records from the table
func (r *tileLock) All() (all map[string]string, err error) {
	all = map[string]string{}
	keys, vals, err := r.tileDB.GetRanged(nil, 0, false)
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
	r.tileDB.Close()
	r.userDB.Close()
}
