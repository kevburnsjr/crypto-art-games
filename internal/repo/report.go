package repo

import (
	"encoding/binary"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Report interface {
	Insert(report *entity.Report) (err error)
	All() (reports []*entity.Report, err error)
	Sweep(t time.Time) (s int, n int, err error)
	Clear(targetID uint16) (err error)
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
func (r *report) Insert(report *entity.Report) (err error) {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint16(idBytes[0:2], report.TargetID)
	binary.BigEndian.PutUint16(idBytes[2:4], report.BoardID)
	binary.BigEndian.PutUint16(idBytes[4:6], report.Timecode)
	binary.BigEndian.PutUint16(idBytes[6:8], report.UserID)
	val := make([]byte, 8)
	binary.BigEndian.PutUint32(val[0:4], report.Date)
	binary.BigEndian.PutUint32(val[4:8], report.FrameDate)
	_, err = r.db.Put(idBytes, "", append(val, []byte(report.Reason)...))
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
		if len(keys[i]) != 8 {
			continue
		}
		reports = append(reports, &entity.Report{
			TargetID:  binary.BigEndian.Uint16(keys[i][0:2]),
			BoardID:   binary.BigEndian.Uint16(keys[i][2:4]),
			Timecode:  binary.BigEndian.Uint16(keys[i][4:6]),
			UserID:    binary.BigEndian.Uint16(keys[i][6:8]),
			Date:      binary.BigEndian.Uint32(val[0:4]),
			FrameDate: binary.BigEndian.Uint32(val[4:8]),
			Reason:    string(val[8:]),
		})
	}
	return
}

// Clear
func (r *report) Clear(targetID uint16) (err error) {
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes[0:2], targetID)
	iter, err := r.db.PrefixIterator(idBytes)
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		err = r.db.Delete([]byte(iter.Key()), "")
		if err != nil {
			break
		}
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
