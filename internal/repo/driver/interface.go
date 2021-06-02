package driver

type DB interface {
	Get(key []byte) (version string, value []byte, err error)
	Put(key []byte, prev string, value []byte) (version string, err error)
	Delete(key []byte, prev string) (err error)
	Has(key []byte) (exists bool, err error)
	PutRanged(id, date string, value []byte) (err error)
	GetRanged(start []byte, limit int, reverse bool) (keys [][]byte, values [][]byte, err error)
	Iterator() (Iterator, error)
	PrefixIterator(prefix []byte) (Iterator, error)
	Close() error
	/*
		Batch() Batch
		RangeIterator(start, limit string) Iterator
		OpenTransaction() (Transaction, error)
	*/

}

type Iterator interface {
	Valid() bool
	Seek(key []byte) bool
	First() bool
	Last() bool
	Prev() bool
	Next() bool
	Key() []byte
	Value() []byte
	Release()
	Error() error
}

/*
type Batch interface {
	Put(key, value string)
	Delete(key string)
	Write() error
}

type Transaction interface {
	Put(key, value string)
	Delete(key string)
	Commit() error
	Discard()
}
*/
