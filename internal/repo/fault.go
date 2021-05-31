package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Fault interface {
	Insert(errType string, userID uint16, userAgent string, t time.Time) (err error)
	All() (faults []*entity.Fault, err error)
	Sweep(t time.Time) (s int, n int, err error)
}

// NewFault returns a Fault repo instance
func NewFault(cfg config.KeyValueStore) (r *fault, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &fault{
		db: db,
	}, nil
}

type fault struct {
	db driver.DB
}

// Insert inserts a fault
func (r *fault) Insert(errType string, userID uint16, userAgent string, t time.Time) (err error) {
	idBytes := make([]byte, 6)
	binary.BigEndian.PutUint32(idBytes[0:4], uint32(t.Truncate(time.Minute).Unix()))
	binary.BigEndian.PutUint16(idBytes[4:6], userID)
	idBytes = append(idBytes, []byte(errType)...)
	_, err = r.db.Put(idBytes, "", []byte(userAgent))
	return
}

// All fetches all faults
func (r *fault) All() (faults []*entity.Fault, err error) {
	var start = make([]byte, 2)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, val := range vals {
		if len(keys[i]) != 4 {
			continue
		}
		faults = append(faults, &entity.Fault{
			ErrType:   string(keys[i][6:]),
			UserID:    binary.BigEndian.Uint16(keys[i][4:6]),
			Date:      time.Unix(int64(binary.BigEndian.Uint32(val[0:4])), 0),
			UserAgent: string(val),
		})
	}
	return
}

// Sweep deletes all faults older than a given timestamp returning number scanned and number deleted
func (r *fault) Sweep(t time.Time) (s int, n int, err error) {
	keys, _, err := r.db.GetRanged([]byte(nil), 0, false)
	if err != nil {
		return
	}
	for _, key := range keys {
		if time.Unix(int64(binary.BigEndian.Uint32(key[0:4])), 0).Before(t) {
			if err = r.db.Delete(key, ""); err != nil {
				return
			}
			n++
		}
		s++
	}
	return
}

// Close closes a database connection
func (r *fault) Close() {
	r.db.Close()
}
