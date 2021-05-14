package config

type Repos struct {
	Global *KeyValueStore `yaml:"global"`

	User  KeyValueStore `yaml:"user"`
	Frame KeyValueStore `yaml:"frame"`
	FrameLock KeyValueStore `yaml:"frameLock"`
	TileHistory KeyValueStore `yaml:"tileHistory"`
	UserFrameHistory KeyValueStore `yaml:"userFrameHistory"`
}

type KeyValueStore struct {
	Cache      *RepoCache  `yaml:"cache"`
	InMemoryDB *InMemoryDB `yaml:"inmemorydb"`
	LevelDB    *LevelDB    `yaml:"leveldb"`
}

func (k *KeyValueStore) Override(t *KeyValueStore) {
	if t.InMemoryDB != nil && k.InMemoryDB != nil {
		k.InMemoryDB.Override(t.InMemoryDB)
	}
	if t.LevelDB != nil && k.LevelDB != nil {
		k.LevelDB.Override(t.LevelDB)
	}
}

type InMemoryDB struct {
	Size int `yaml:"size"`
}

func (k *InMemoryDB) Override(t *InMemoryDB) {
	if k.Size > 0 && t.Size == 0 {
		t.Size = k.Size
	}
}

type LevelDB struct {
	Path string `yaml:"path"`
}

func (k *LevelDB) Override(t *LevelDB) {
	if len(k.Path) > 0 && len(t.Path) == 0 {
		t.Path = k.Path
	}
}

type RepoCache struct {
	Enabled bool `yaml:"enabled"`
	Size    int  `yaml:"size"`
}
