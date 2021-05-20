package repo

import (
	"encoding/binary"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type User interface {
	Find(user *entity.User) (userID uint16, found bool, err error)
	FindOrInsert(user *entity.User) (userID uint16, inserted bool, err error)
	Update(user *entity.User) (u *entity.User, err error)
	Since(userIdx, generation uint16) (users []*entity.User, userIds []uint16, err error)
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

// Find retrieves a user
func (r *user) Find(user *entity.User) (userID uint16, found bool, err error) {
	_, bytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err == errors.RepoItemNotFound {
		err = nil
		return
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint16(bytes)
		found = true
	}
	return
}

// FindOrInsert inserts a user or returns the user's existing ID
func (r *user) FindOrInsert(user *entity.User) (userID uint16, inserted bool, err error) {
	_, bytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err == errors.RepoItemNotFound {
		userID, err = r.Insert(user)
		inserted = true
		return
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

	user.UserID = userID

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

func (r *user) Update(user *entity.User) (u *entity.User, err error) {
	_, idBytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err != nil {
		return
	}
	userVers, userBytes, err := r.db.Get(idBytes)
	if err != nil {
		return
	}

	u = entity.UserFromJson(userBytes)
	u.User = user.User
	if user.Policy {
		u.Policy = user.Policy
	}

	_, err = r.db.Put(idBytes, userVers, u.ToJson())
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

// Since returns new users
func (r *user) Since(userIdx, generation uint16) (users []*entity.User, userIds []uint16, err error) {
	var start = make([]byte, 2)
	binary.BigEndian.PutUint16(start, userIdx)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, b := range vals {
		if len(keys[i]) != 2 {
			continue
		}
		userIds = append(userIds, binary.BigEndian.Uint16(keys[i]))
		users = append(users, entity.UserFromJson(b))
	}
	return
}

// Close closes a database connection
func (r *user) Close() {
	r.db.Close()
}
