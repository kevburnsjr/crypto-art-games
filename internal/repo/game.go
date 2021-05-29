package repo

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/entity"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
	"github.com/kevburnsjr/crypto-art-games/internal/repo/driver"
)

type Game interface {
	Version() (v uint64, err error)
	ActiveSeries() (all entity.SeriesList, err error)
	AllSeries() (all entity.SeriesList, err error)
	InsertSeries(data string) (err error)
	UpdateSeries(id, data string) (err error)
	FindSeries(id string) (res *entity.Series, err error)
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
	iter, err := r.db.PrefixIterator("series-")
	if err != nil {
		return
	}
	for iter.Next() {
		c := entity.SeriesFromJson(iter.Value()[16:])
		if c == nil || !c.Active {
			continue
		}
		all = append(all, c)
	}
	return
}

// AllSeries retrieves a list of all seriess
func (r *game) AllSeries() (all entity.SeriesList, err error) {
	iter, err := r.db.PrefixIterator("series-")
	if err != nil {
		return
	}
	for iter.Next() {
		c := entity.SeriesFromJson(iter.Value()[16:])
		if c == nil {
			continue
		}
		all = append(all, c)
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
func (r *game) InsertSeries(data string) (err error) {
	var id uint16
	_, idBytes, err := r.db.Get([]byte("_id"))
	if err == errors.RepoItemNotFound {
		id = uint16(1)
	} else if err != nil {
		return
	} else {
		id = binary.BigEndian.Uint16(idBytes)
		id++
	}
	_, err = r.db.Put([]byte(fmt.Sprintf("series-%04x", id)), "", []byte(data))
	return
}

// UpdateSeries updates a new series
func (r *game) UpdateSeries(id, data string) (err error) {
	_, err = r.db.Put([]byte("series-"+id), "", []byte(data))
	return
}

// Close closes a database connection
func (r *game) Close() {
	r.db.Close()
}
