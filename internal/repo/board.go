package repo

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Board interface {
	Find(boardId, timecode uint16) (frame *entity.Frame, err error)
	Insert(boardId uint16, frame *entity.Frame, t time.Time) (timecode uint16, err error)
	Since(boardId, generation, timecode uint16) (frames []*entity.Frame, err error)
}

// NewBoard returns an Frame repo instance
func NewBoard(cfg config.KeyValueStore) (r *board, err error) {
	var dbFactory func(uint16) (driver.DB, error)
	if cfg.LevelDB != nil {
		dbFactory = func(boardId uint16) (driver.DB, error) {
			dbcfg := *cfg.LevelDB
			dbcfg.Path += fmt.Sprintf("-%04x", boardId)
			return driver.NewLevelDB(dbcfg)
		}
	}
	if err != nil || dbFactory == nil {
		return
	}
	return &board{
		dbMap:     map[uint16]driver.DB{},
		dbFactory: dbFactory,
	}, nil
}

type board struct {
	dbMap     map[uint16]driver.DB
	dbFactory func(uint16) (driver.DB, error)
}

func (r *board) db(boardId uint16) (driver.DB, error) {
	if db, ok := r.dbMap[boardId]; ok {
		return db, nil
	}
	var err error
	r.dbMap[boardId], err = r.dbFactory(boardId)
	return r.dbMap[boardId], err
}

// Find returns a frame by timecode
func (r *board) Find(boardId, timecode uint16) (frame *entity.Frame, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	tcBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tcBytes, timecode)
	_, frameBytes, err := db.Get(tcBytes)
	frame = entity.FrameFromBytes(frameBytes)
	return
}

// Insert inserts a frame
func (r *board) Insert(boardId uint16, frame *entity.Frame, t time.Time) (timecode uint16, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	tcVers, tcBytes, err := db.Get([]byte("_timecode"))
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

	err = r.timeCheck(boardId, frame, t)
	if err != nil {
		return
	}

	_, err = db.Put(tcBytes, "", frame.ToBytes())
	if err != nil {
		return
	}

	_, err = db.Put([]byte("_timecode"), tcVers, tcBytes)
	return
}

// Checkpoint encoded timestamps
func (r *board) timeCheck(boardId uint16, frame *entity.Frame, t time.Time) (err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	var timecheck uint32
	chkVers, chkBytes, err := db.Get([]byte("_timecheck"))
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
		_, err = db.Put([]byte("_timecheck"), chkVers, chkBytes)
		frame.SetTimestamp(0)
		frame.SetTimecheck(timecheck)
	} else {
		frame.SetTimestamp(uint16(timestamp / 60))
	}

	return
}

// Since inserts all frames since timecode
func (r *board) Since(boardId, generation, timecode uint16) (frames []*entity.Frame, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	var start = make([]byte, 2)
	binary.BigEndian.PutUint16(start, timecode)
	keys, vals, err := db.GetRanged(start, 0, false)
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
func (r *board) Close() {
	for _, db := range r.dbMap {
		db.Close()
	}
}
