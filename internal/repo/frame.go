package repo

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Frame interface {
	Insert(frame *entity.Frame, t time.Time) (timecode uint16, err error)
	Since(boardId, generation, timecode uint16) (frames []*entity.Frame, err error)
}

// NewFrame returns an Frame repo instance
func NewFrame(cfg config.KeyValueStore) (r *frame, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &frame{
		db: db,
	}, nil
}

type frame struct {
	db driver.DB
}

// Insert inserts a frame
func (r *frame) Insert(frame *entity.Frame, t time.Time) (timecode uint16, err error) {
	tcVers, tcBytes, err := r.db.Get([]byte("_timecode"))
	if err == errors.RepoItemNotFound {
		timecode = uint16(0)
	} else if err != nil {
		return
	} else {
		timecode = binary.BigEndian.Uint16(tcBytes)
		timecode++
	}
	frame.SetTimecode(timecode)

	tcBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(tcBytes, timecode)

	err = r.timeCheck(frame, t)
	if err != nil {
		return
	}

	_, err = r.db.Put(tcBytes, "", frame.ToBytes())
	if err != nil {
		return
	}

	_, err = r.db.Put([]byte("_timecode"), tcVers, tcBytes)
	return
}

// Checkpoint encoded timestamps
func (r *frame) timeCheck(frame *entity.Frame, t time.Time) (err error) {
	var timecheck uint32
	chkVers, chkBytes, err := r.db.Get([]byte("_timecheck"))
	if err == errors.RepoItemNotFound {
		timecheck = 0
		err = nil
	} else if err != nil {
		return
	} else {
		timecheck = binary.BigEndian.Uint32(chkBytes)
	}
	var timestamp = t.Truncate(60*time.Second).Unix() - int64(timecheck)
	if timecheck == 0 || timestamp > math.MaxUint16-1 {
		timecheck = uint32(t.Truncate(60*time.Second).Unix()) - 60
		chkBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(chkBytes, timecheck)
		_, err = r.db.Put([]byte("_timecheck"), chkVers, chkBytes)
		frame.SetTimestamp(0)
		frame.SetTimecheck(timecheck)
	} else {
		frame.SetTimestamp(uint16(timestamp / 60))
	}

	return
}

// Since inserts all frames since timecode
func (r *frame) Since(boardId, generation, timecode uint16) (frames []*entity.Frame, err error) {
	var start = make([]byte, 2)
	binary.BigEndian.PutUint16(start, timecode)
	keys, vals, err := r.db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, b := range vals {
		if len(keys[i]) != 2 {
			continue
		}
		frame := &entity.Frame{
			Data: b,
		}
		frames = append(frames, frame)
	}
	return
}

// Close closes a database connection
func (r *frame) Close() {
	r.db.Close()
}
