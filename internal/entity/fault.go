package entity

import (
	"encoding/json"
	"time"
)

type Fault struct {
	ErrType   string    `json:"type"`
	UserID    uint32    `json:"userID"`
	Date      time.Time `json:"date"`
	UserAgent string    `json:"userAgent"`
}

func (u *Fault) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}
