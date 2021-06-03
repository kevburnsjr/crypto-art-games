package repo

import (
	"encoding/binary"
	"fmt"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Board interface {
	Find(boardId uint16, timecode uint32) (frame *entity.Frame, err error)
	Insert(boardId uint16, frame *entity.Frame) (err error)
	Since(boardId uint16, timecode uint32) (frames []*entity.Frame, err error)
	Update(boardId uint16, f *entity.Frame) (err error)
	DeleteUserFramesAfter(boardId uint16, targetID, timecode uint32) (n int, err error)
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
func (r *board) Find(boardId uint16, timecode uint32) (frame *entity.Frame, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	tcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tcBytes, timecode)
	_, frameBytes, err := db.Get(tcBytes)
	frame = entity.FrameFromBytes(frameBytes)
	return
}

// Insert inserts a frame
func (r *board) Insert(boardId uint16, f *entity.Frame) (err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	_, err = db.Put(f.ID(), "", f.ToBytes())
	if err != nil {
		return
	}
	return
}

// Update updates a frame
func (r *board) Update(boardId uint16, f *entity.Frame) (err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	_, err = db.Put(f.ID(), "", f.ToBytes())
	return
}

// Since inserts all frames since timecode
func (r *board) Since(boardId uint16, timecode uint32) (frames []*entity.Frame, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	var start = make([]byte, 4)
	binary.BigEndian.PutUint32(start, timecode-timecode%256)
	keys, vals, err := db.GetRanged(start, 0, false)
	if err != nil {
		return
	}
	for i, b := range vals {
		if len(keys[i]) != 4 {
			continue
		}
		frame := &entity.Frame{
			Data: b,
		}
		frames = append(frames, frame)
	}
	return
}

// DeleteUserFramesAfter removes a users' contributions to the board
func (r *board) DeleteUserFramesAfter(boardId uint16, targetID, timecode uint32) (n int, err error) {
	db, err := r.db(boardId)
	if err != nil {
		return
	}
	iter, err := db.Iterator()
	if err != nil {
		return
	}
	defer iter.Release()
	var f = &entity.Frame{}
	var start = make([]byte, 4)
	binary.BigEndian.PutUint32(start, timecode-timecode%256)
	for iter.Seek(start); iter.Valid(); iter.Next() {
		if len(iter.Key()) != 4 {
			continue
		}
		f.Data = iter.Value()[16:]
		if f.UserID() == targetID {
			f.SetDeleted(true)
			if err = r.Update(boardId, f); err != nil {
				return
			}
			n++
		}
	}
	return
}

// Close closes a database connection
func (r *board) Close() {
	for _, db := range r.dbMap {
		db.Close()
	}
}
