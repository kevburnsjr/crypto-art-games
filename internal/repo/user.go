package repo

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type User interface {
	Find(user *entity.User) (userID uint32, found bool, err error)
	FindByUserID(userID uint32) (user *entity.User, err error)
	FindByUserIDStr(userID string) (user *entity.User, err error)
	FindOrInsert(user *entity.User) (userID uint32, inserted bool, err error)
	Update(user *entity.User) (err error)
	UpdateProfile(user *entity.User) (u *entity.User, err error)
	Since(userIdx uint32) (users []*entity.User, userIds []uint32, err error)
	Consume(user *entity.User, boardId uint16) (err error)
	Credit(user *entity.User, boardId uint16) (err error)
	All() (all []*entity.User, err error)
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
func (r *user) Find(user *entity.User) (userID uint32, found bool, err error) {
	_, idBytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err == errors.RepoItemNotFound {
		err = nil
		return
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint32(idBytes)
		found = true
	}

	_, userBytes, err := r.db.Get(idBytes)
	if err != nil {
		return
	}
	err = json.Unmarshal(userBytes, user)

	return
}

// FindByUserIDStr retrieves a user with a userid string
func (r *user) FindByUserIDStr(userID string) (user *entity.User, err error) {
	idInt, err := strconv.Atoi(userID)
	if err != nil {
		return
	}
	return r.FindByUserID(uint32(idInt))
}

// FindByUserID retrieves a user
func (r *user) FindByUserID(userID uint32) (user *entity.User, err error) {
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, userID)
	_, userBytes, err := r.db.Get(idBytes)
	if err != nil {
		return
	}
	user = &entity.User{}
	err = json.Unmarshal(userBytes, user)

	return
}

// FindOrInsert inserts a user or returns the user's existing ID
func (r *user) FindOrInsert(user *entity.User) (userID uint32, inserted bool, err error) {
	_, bytes, err := r.db.Get([]byte("twitch-" + user.ID))
	if err == errors.RepoItemNotFound {
		userID, err = r.Insert(user)
		inserted = true
		err = nil
		return
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint32(bytes)
	}
	return
}

func (r *user) Insert(user *entity.User) (userID uint32, err error) {
	idVers, idBytes, err := r.db.Get([]byte("_id"))
	if err == errors.RepoItemNotFound {
		userID = uint32(1)
		err = nil
	} else if err != nil {
		return
	} else {
		userID = binary.BigEndian.Uint32(idBytes)
		userID++
	}
	idBytes = make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, userID)

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

func (r *user) Update(user *entity.User) (err error) {
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, user.UserID)
	_, err = r.db.Put(idBytes, "", user.ToJson())
	return
}

func (r *user) UpdateProfile(user *entity.User) (u *entity.User, err error) {
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

func (r *user) Consume(user *entity.User, boardId uint16) (err error) {
	if user.Buckets == nil {
		user.Buckets = map[uint16]*entity.UserBucket{}
	}
	bucket, ok := user.Buckets[boardId]
	if !ok {
		user.Buckets[boardId] = entity.NewUserBucket(time.Now())
	}
	if !bucket.Consume(1, time.Now()) {
		return fmt.Errorf("Insufficient tile credits")
	}
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, user.UserID)
	userVers, _, err := r.db.Get(idBytes)
	if err != nil {
		return
	}
	_, err = r.db.Put(idBytes, userVers, user.ToJson())
	if err != nil {
		return
	}

	return
}

func (r *user) Credit(user *entity.User, boardId uint16) (err error) {
	if user.Buckets == nil {
		user.Buckets = map[uint16]*entity.UserBucket{}
	}
	bucket, ok := user.Buckets[boardId]
	if !ok {
		user.Buckets[boardId] = entity.NewUserBucket(time.Now())
	}
	bucket.Credit(1, time.Now())
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, user.UserID)
	userVers, _, err := r.db.Get(idBytes)
	if err != nil {
		return
	}
	_, err = r.db.Put(idBytes, userVers, user.ToJson())
	if err != nil {
		return
	}

	return
}

// All returns all records from the table
func (r *user) All() (all []*entity.User, err error) {
	iter, err := r.db.PrefixIterator(nil)
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		c := entity.UserFromJson(iter.Value()[16:])
		if c != nil {
			all = append(all, c)
		}
	}
	return
}

// Since returns new users
func (r *user) Since(userIdx uint32) (users []*entity.User, userIds []uint32, err error) {
	var start = make([]byte, 4)
	binary.BigEndian.PutUint32(start, userIdx)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, b := range vals {
		if len(keys[i]) != 4 || bytes.Compare(start, keys[i]) == 0 {
			continue
		}
		userIds = append(userIds, binary.BigEndian.Uint32(keys[i]))
		users = append(users, entity.UserFromJson(b))
	}
	return
}

// Close closes a database connection
func (r *user) Close() {
	r.db.Close()
}
