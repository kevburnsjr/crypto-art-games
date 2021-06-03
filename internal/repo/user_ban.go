package repo

import (
	"bytes"
	"encoding/binary"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type UserBan interface {
	Insert(userBan *entity.UserBan) (err error)
	Since(index uint32) (userBans []*entity.UserBan, err error)
	All() (all []*entity.UserBan, err error)
}

// NewUserBan returns a UserBan repo instance
func NewUserBan(cfg config.KeyValueStore) (r *userBan, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &userBan{
		db: db,
	}, nil
}

type userBan struct {
	db driver.DB
}

// Insert inserts a userBan
func (r *userBan) Insert(userBan *entity.UserBan) (err error) {
	var id uint32
	idVers, idBytes, err := r.db.Get([]byte("_id"))
	if err == errors.RepoItemNotFound {
		id = uint32(1)
	} else if err != nil {
		return
	} else {
		id = binary.BigEndian.Uint32(idBytes)
		id++
	}
	idBytes = make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, id)

	userBan.ID = id

	_, err = r.db.Put(idBytes, "", userBan.ToJson())
	if err != nil {
		return
	}

	_, err = r.db.Put([]byte("_id"), idVers, idBytes)
	return
}

// Since inserts all userBans since timecode
func (r *userBan) Since(id uint32) (userBans []*entity.UserBan, err error) {
	var start = make([]byte, 4)
	binary.BigEndian.PutUint32(start, id)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, b := range vals {
		if len(keys[i]) != 2 || bytes.Compare(start, keys[i]) == 0 {
			continue
		}
		userBans = append(userBans, entity.UserBanFromJson(b))
	}
	return
}

// All returns all records from the table
func (r *userBan) All() (all []*entity.UserBan, err error) {
	iter, err := r.db.PrefixIterator(nil)
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		c := entity.UserBanFromJson(iter.Value()[16:])
		if c != nil {
			all = append(all, c)
		}
	}
	return
}

// Close closes a database connection
func (r *userBan) Close() {
	r.db.Close()
}
