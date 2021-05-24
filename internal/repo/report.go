package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Report interface {
	Insert(userID, timecode uint16, reason string, t time.Time) (err error)
	All() (reports []*entity.Report, err error)
	Sweep(t time.Time) (s int, n int, err error)
}

// NewReport returns a Report repo instance
func NewReport(cfg config.KeyValueStore) (r *report, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &report{
		db: db,
	}, nil
}

type report struct {
	db driver.DB
}

// Insert inserts a report
func (r *report) Insert(userID, timecode uint16, reason string, t time.Time) (err error) {
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(idBytes[0:2], timecode)
	binary.BigEndian.PutUint16(idBytes[2:4], userID)
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val[0:4], uint32(t.Unix()))
	_, err = r.db.Put(idBytes, "", append(val, []byte(reason)...))
	return
}

// All fetches all reports
func (r *report) All() (reports []*entity.Report, err error) {
	var start = make([]byte, 2)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, val := range vals {
		if len(keys[i]) != 4 {
			continue
		}
		reports = append(reports, &entity.Report{
			Timecode: binary.BigEndian.Uint16(keys[i][0:2]),
			UserID:   binary.BigEndian.Uint16(keys[i][2:4]),
			Date:     time.Unix(int64(binary.BigEndian.Uint32(val[0:4])), 0),
			Reason:   string(val[4:]),
		})
	}
	return
}

// Sweep deletes all reports older than a given timestamp returning number scanned and number deleted
func (r *report) Sweep(t time.Time) (s int, n int, err error) {
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
func (r *report) Close() {
	r.db.Close()
}
