package repo

import (
	"encoding/binary"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Frame interface {
	Insert(frame *entity.Frame) (timecode uint16, err error)
	Since(timecode uint16) (frames []*entity.Frame, err error)
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
func (r *frame) Insert(frame *entity.Frame) (timecode uint16, err error) {
	tcVers, tcBytes, err := r.db.Get([]byte("_timecode"))
	if err == errors.RepoItemNotFound {
		timecode = uint16(0)
	} else if err != nil {
		return
	} else {
		timecode = binary.BigEndian.Uint16(tcBytes)
		timecode++
	}
	tcBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(tcBytes, timecode)

	_, err = r.db.Put(tcBytes, "", frame.Data)
	if err != nil {
		return
	}

	frame.Timecode = timecode

	_, err = r.db.Put([]byte("_timecode"), tcVers, tcBytes)
	return
}

// Since inserts all frames since timecode
func (r *frame) Since(timecode uint16) (frames []*entity.Frame, err error) {
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
			Timecode: binary.BigEndian.Uint16(keys[i]),
			Data:     b,
		}
		frames = append(frames, frame)
	}
	return
}

// Close closes a database connection
func (r *frame) Close() {
	r.db.Close()
}
