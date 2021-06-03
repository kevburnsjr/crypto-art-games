package repo

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Game interface {
	Version() (v uint64, err error)
	ActiveSeries() (all entity.SeriesList, err error)
	AllSeries() (all entity.SeriesList, err error)
	InsertSeries(series *entity.Series) (err error)
	UpdateSeries(id string, series *entity.Series) (err error)
	FindSeries(id string) (res *entity.Series, err error)
	FindActiveBoard(boardId uint16) (board *entity.Board, err error)
}

// NewGame returns an Game repo instance
func NewGame(cfg config.KeyValueStore) (r *game, err error) {
	var db driver.DB
	if cfg.LevelDB != nil {
		db, err = driver.NewLevelDB(*cfg.LevelDB)
	}
	if err != nil || db == nil {
		return
	}
	return &game{
		db: db,
	}, nil
}

type game struct {
	db driver.DB
}

// Version retrieves or sets the game version
func (r *game) Version() (v uint64, err error) {
	_, vBytes, err := r.db.Get([]byte("_v"))
	if err == errors.RepoItemNotFound {
		vBytes = make([]byte, 8)
		rand.Read(vBytes)
		_, err = r.db.Put([]byte("_v"), "", vBytes)
		err = nil
	} else if err != nil {
		return
	} else {
		v = binary.BigEndian.Uint64(vBytes)
	}

	return
}

// ActiveSeries retrieves a list of all active series
func (r *game) ActiveSeries() (all entity.SeriesList, err error) {
	iter, err := r.db.PrefixIterator([]byte("series-"))
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		c := entity.SeriesFromJson(iter.Value()[16:])
		if c == nil || c.Active == 0 || c.Active > uint32(time.Now().Unix()) {
			continue
		}
		all = append(all, c)
	}
	return
}

// AllSeries retrieves a list of all series
func (r *game) AllSeries() (all entity.SeriesList, err error) {
	iter, err := r.db.PrefixIterator([]byte("series-"))
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		c := entity.SeriesFromJson(iter.Value()[16:])
		if c == nil {
			continue
		}
		all = append(all, c)
	}
	return
}

// FindActiveBoard retrieves an active board by id
func (r *game) FindActiveBoard(boardId uint16) (board *entity.Board, err error) {
	iter, err := r.db.PrefixIterator([]byte("series-"))
	if err != nil {
		return
	}
	defer iter.Release()
	for iter.Next() {
		s := entity.SeriesFromJson(iter.Value()[16:])
		if s == nil || s.Active == 0 || s.Active > uint32(time.Now().Unix()) {
			continue
		}
		for _, b := range s.Boards {
			if b.ID == boardId {
				b.Created = s.Created
				return &b, nil
			}
		}
	}
	return
}

// FindSeries inserts a new series
func (r *game) FindSeries(id string) (res *entity.Series, err error) {
	_, b, err := r.db.Get([]byte("series-" + id))
	res = entity.SeriesFromJson(b)
	return
}

// InsertSeries inserts a new series
func (r *game) InsertSeries(series *entity.Series) (err error) {
	var id uint16
	idVers, idBytes, err := r.db.Get([]byte("_id"))
	if err == errors.RepoItemNotFound {
		id = uint16(1)
	} else if err != nil {
		return
	} else {
		id = binary.BigEndian.Uint16(idBytes)
		id++
	}
	idBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, id)
	_, err = r.db.Put([]byte("_id"), idVers, idBytes)
	if err != nil {
		return
	}
	series.ID = id
	_, err = r.db.Put([]byte(fmt.Sprintf("series-%04x", id)), "", series.ToJson())
	return
}

// UpdateSeries updates a new series
func (r *game) UpdateSeries(id string, series *entity.Series) (err error) {
	i, _ := strconv.Atoi(id)
	series.ID = uint16(i)
	_, err = r.db.Put([]byte("series-"+id), "", series.ToJson())
	return
}

// Close closes a database connection
func (r *game) Close() {
	r.db.Close()
}
