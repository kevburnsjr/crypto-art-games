package driver

import (
	"crypto/sha256"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	leveldbErr "github.com/syndtr/goleveldb/leveldb/errors"
	// "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/errors"
)

var leveldbInstances = map[string]*leveldbDriver{}

func NewLevelDB(cfg config.LevelDB) (w *leveldbDriver, err error) {
	w, ok := leveldbInstances[cfg.Path]
	if !ok {
		db, err := leveldb.OpenFile(cfg.Path, nil)
		if err != nil {
			return nil, errors.RepoDBUnavailable
		}
		w = &leveldbDriver{db}
		leveldbInstances[cfg.Path] = w
	}
	return
}

type leveldbDriver struct {
	db *leveldb.DB
}

func (w *leveldbDriver) Get(key []byte) (version string, value []byte, err error) {
	value, err = w.db.Get(key, nil)
	if err == leveldbErr.ErrNotFound {
		err = errors.RepoItemNotFound
		return
	}
	if err == nil && len(value) > 16 {
		version = string(value[:16])
		value = value[16:]
	}
	return
}

func (w *leveldbDriver) Put(key []byte, prev string, value []byte) (version string, err error) {
	version = fmt.Sprintf("%x", sha256.Sum256(value))[:16]
	v, _, err := w.Get(key)
	if prev != "" && err != nil {
		return
	}
	if prev != "" && v != prev {
		println(v, prev)
		err = errors.RepoItemVersionConflict
		return
	}
	err = w.db.Put(key, append([]byte(version), value...), nil)
	return
}

func (w *leveldbDriver) Delete(key []byte, prev string) (err error) {
	v, _, err := w.Get(key)
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
	err = w.db.Delete([]byte(key), nil)
	return
}

func (w *leveldbDriver) Has(key []byte) (exists bool, err error) {
	exists, err = w.db.Has(key, nil)
	return
}

func (w *leveldbDriver) PutRanged(id, date string, value []byte) (err error) {
	key := id + "-" + date
	w.db.Put([]byte(key), value, nil)
	return
}

func (w *leveldbDriver) GetRanged(start []byte, limit int, reverse bool) (keys [][]byte, values [][]byte, err error) {
	iter := w.db.NewIterator(&util.Range{start, nil}, nil)
	defer iter.Release()
	for iter.First(); iter.Valid(); iter.Next() {
		keys = append(keys, []byte(string(iter.Key())))
		values = append(values, []byte(string(iter.Value()[16:])))
		if limit > 0 && len(values) >= limit {
			break
		}
	}
	if reverse {
		// NOT TODO - support reverse (meh)
	}
	return
}

func (w *leveldbDriver) Close() error {
	return w.db.Close()
}

/*
func (w *leveldbDriver) Batch() Batch {
	return leveldb_batch{w.db, new(leveldb.Batch)}
}
func (w *leveldbDriver) PrefixIterator(prefix string) Iterator {
	return leveldb_iterator{w.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)}
}
func (w *leveldbDriver) RangeIterator(start, limit string) Iterator {
	return leveldb_iterator{w.db.NewIterator(&util.Range{Start: []byte(start), Limit: []byte(limit)}, nil)}
}
func (w *leveldbDriver) OpenTransaction() (Transaction, error) {
	transaction, err := w.db.OpenTransaction()
	if err != nil {
		return leveldb_transaction{}, err
	}
	return leveldb_transaction{transaction}, nil
}

type leveldb_batch struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

func (b leveldb_batch) Put(key string, value []byte) {
	b.batch.Put([]byte(key), value)
	return
}
func (b leveldb_batch) Delete(key string) {
	b.batch.Delete([]byte(key))
	return
}
func (b leveldb_batch) Write() error {
	return b.db.Write(b.batch, nil)
}

type leveldb_iterator struct {
	iter iterator.Iterator
}

func (i leveldb_iterator) Seek(key string) bool {
	return i.iter.Seek([]byte(key))
}
func (i leveldb_iterator) Next() bool {
	return i.iter.Next()
}
func (i leveldb_iterator) Key() string {
	return string(i.iter.Key())
}
func (i leveldb_iterator) Value() []byte {
	return i.iter.Value()
}
func (i leveldb_iterator) Release() {
	i.iter.Release()
	return
}
func (i leveldb_iterator) Error() error {
	return i.iter.Error()
}

type leveldb_transaction struct {
	transaction *leveldb.Transaction
}

func (t leveldb_transaction) Put(key string, value []byte) {
	t.transaction.Put([]byte(key), value, nil)
	return
}
func (t leveldb_transaction) Delete(key string) {
	t.transaction.Delete([]byte(key), nil)
	return
}
func (t leveldb_transaction) Commit() error {
	return t.transaction.Commit()
}
func (t leveldb_transaction) Discard() {
	t.transaction.Discard()
	return
}
*/
