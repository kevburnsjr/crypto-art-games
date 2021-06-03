package entity

import (
	"encoding/json"
	"time"
)

type Love struct {
	BoardID  uint16    `json:"boardID"`
	Timecode uint32    `json:"timecode"`
	UserID   uint32    `json:"userID"`
	Date     time.Time `json:"date"`
}

type LoveDto struct {
	Love
	Type string `json:"type"`
}

func (u *Love) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *Love) ToDto() []byte {
	b, _ := json.Marshal(LoveDto{*u, "love"})
	return b
}
