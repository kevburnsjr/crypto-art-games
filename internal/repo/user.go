package repo

import (
	"encoding/binary"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type User interface {
	FindOrInsert(user *entity.User) (userID uint16, err error)
}

// NewUser returns an User repo instance
func NewUser(cfg config.KeyValueStore) (r *user, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &user{
		db: db,
	}, nil
}

type user struct {
	db driver.DB
}

// Find retrieves an user
func (r *user) FindOrInsert(user *entity.User) (userID uint16, err error) {
	_, bytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err == errors.RepoItemNotFound {
		return r.Insert(user)
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint16(bytes)
	}
	return
}

func (r *user) Insert(user *entity.User) (userID uint16, err error) {
	idVers, idBytes, err := r.db.Get([]byte("_id"))
	if err == errors.RepoItemNotFound {
		userID = uint16(0)
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint16(idBytes)
		userID++
	}
	idBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, userID)

	_, err = r.db.Put([]byte("_id"), idVers, idBytes)
	if err != nil {
		return
	}

	_, err = r.db.Put(idBytes, "", user.ToJson())
	if err != nil {
		return
	}

	_, err = r.db.Put([]byte("twitch-"+user.ID), "", idBytes)
	if err != nil {
		return
	}

	return
}

// All returns all records from the table
func (r *user) All() (all map[string]string, err error) {
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
func (r *user) Close() {
	r.db.Close()
}
