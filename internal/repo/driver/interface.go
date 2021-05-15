package driver

type DB interface {
	Get(key []byte) (version string, value []byte, err error)
	Put(key []byte, prev string, value []byte) (version string, err error)
	Delete(key []byte, prev string) (err error)
	Has(key []byte) (exists bool, err error)
	PutRanged(id, date string, value []byte) (err error)
	GetRanged(start []byte, limit int, reverse bool) (keys [][]byte, values [][]byte, err error)
	Close() error
	/*
		Batch() Batch
		PrefixIterator(prefix string) Iterator
		RangeIterator(start, limit string) Iterator
		OpenTransaction() (Transaction, error)
	*/

}

/*
type Batch interface {
	Put(key, value string)
	Delete(key string)
	Write() error
}

type Iterator interface {
	Seek(key string) bool
	Next() bool
	Key() string
	Value() string
	Release()
	Error() error
}

type Transaction interface {
	Put(key, value string)
	Delete(key string)
	Commit() error
	Discard()
}
*/
