package driver

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
)

func NewInMemory(cfg config.InMemoryDB) (w *inmemoryDriver, err error) {
	return &inmemoryDriver{sync.Map{}}, nil
}

type inmemoryDriver struct {
	m sync.Map
}

func (d *inmemoryDriver) Get(key []byte) (version string, value []byte, err error) {
	v, ok := d.m.Load(key)
	if !ok || v == nil {
		err = errors.RepoItemNotFound
		return
	}
	value = v.([]byte)
	if len(value) > 16 {
		version = string(value[:16])
		value = value[16:]
	}
	return
}

func (d *inmemoryDriver) Put(key []byte, prev string, value []byte) (version string, err error) {
	version = fmt.Sprintf("%x", sha256.Sum256(value))[:16]
	v, _, err := d.Get(key)
	if err == errors.RepoItemNotFound {
		if prev != "" {
			return
		}
		err = nil
	} else if err != nil {
		return
	}
	if prev != "" && v != prev {
		err = errors.RepoItemVersionConflict
		return
	}
	d.m.Store(string(key), append([]byte(version), value...))
	return
}

func (d *inmemoryDriver) Delete(key []byte, prev string) (err error) {
	v, _, err := d.Get(key)
	if err == errors.RepoItemNotFound {
		return nil
	}
	if err != nil {
		return
	}
	if prev != "" && v != prev {
		err = errors.RepoItemVersionConflict
		return
	}
	d.m.Delete(key)
	return
}

func (d *inmemoryDriver) Has(key []byte) (exists bool, err error) {
	_, exists = d.m.Load(key)
	return
}

func (d *inmemoryDriver) PutRanged(id, date string, value []byte) (err error) {
	key := id + "-" + date
	d.m.Store(key, value)
	return
}

func (d *inmemoryDriver) GetRanged(start []byte, limit int, reverse bool) (keys [][]byte, values [][]byte, err error) {
	startKey := string(start)
	d.m.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		if len(start) == 0 || key > startKey {
			keys = append(keys, k.([]byte))
			values = append(values, v.([]byte))
			if limit > 0 && len(values) >= limit {
				return false
			}
		}
		return true
	})
	if reverse {
		// NOT TODO - support reverse (nobody cares)
	}
	return
}

func (d *inmemoryDriver) Close() error {
	return nil
}
