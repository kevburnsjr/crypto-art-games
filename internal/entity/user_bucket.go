package entity

import (
	"encoding/json"
	"time"
)

type UserBucket struct {
	Size      uint8
	Rate      uint8 // Credit regen rate (seconds per quarter)
	Level     uint8
	Timestamp uint32
}

func NewUserBucket(t time.Time) *UserBucket {
	return &UserBucket{Size: 8, Level: 32, Rate: 15, Timestamp: uint32(t.Unix())}
}

func (b *UserBucket) AdjustLevel(t time.Time) {
	if b == nil {
		return
	}
	var delta = t.Sub(time.Unix(int64(b.Timestamp), 0))
	var levelDelta = int(delta/time.Second) / int(b.Rate)
	if levelDelta+int(b.Level) > int(b.Size*4) {
		b.Level = b.Size * 4
	} else {
		b.Level += uint8(levelDelta)
	}
	b.Timestamp += uint32(levelDelta * int(b.Rate))
}

func (b *UserBucket) Consume(n uint8, t time.Time) bool {
	b.AdjustLevel(t)
	if b.Level < n*4 {
		return false
	}
	b.Level -= n * 4
	return true
}

func (b *UserBucket) Credit(n uint8, t time.Time) {
	b.AdjustLevel(t)
	b.Level = b.Level + n*4
	if b.Level > b.Size*4 {
		b.Level = b.Size * 4
	}
}

func (b *UserBucket) MarshalJSON() ([]byte, error) {
	return json.Marshal([]int{int(b.Size), int(b.Rate), int(b.Level), int(b.Timestamp)})
}

func (b *UserBucket) UnmarshalJSON(in []byte) (err error) {
	var s = make([]int, 4)
	err = json.Unmarshal(in, &s)
	if err != nil {
		return
	}
	*b = UserBucket{uint8(s[0]), uint8(s[1]), uint8(s[2]), uint32(s[3])}
	return
}
