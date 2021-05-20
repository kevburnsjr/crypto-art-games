package entity

import (
	"encoding/json"
	"time"
)

type UserBan struct {
	ID      uint16    `json:"ID"`
	UserID  uint16    `json:"userID"`
	Timeout time.Time `json:"timeout"`
	Band    bool      `json:"ban"`
	UnBan   bool      `json:"unban"`
}

type UserBanDto struct {
	UserBan
	Type string `json:"type"`
}

func (u *UserBan) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *UserBan) ToDto(userID uint16) []byte {
	b, _ := json.Marshal(UserBanDto{*u, "user-ban"})
	return b
}

func UserBanFromJson(b []byte) *UserBan {
	var u UserBan
	err := json.Unmarshal(b, &u)
	if err != nil {
		return nil
	}
	return &u
}
