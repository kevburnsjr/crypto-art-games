package entity

import (
	"encoding/json"
)

type UserBan struct {
	ID       uint16 `json:"ID"`
	ModID    uint16 `json:"modID"`
	TargetID uint16 `json:"TargetID"`
	Since    uint32 `json:"since"`
	Until    uint32 `json:"until"`
	Ban      bool   `json:"ban"`
	UnBan    bool   `json:"unban"`
	Reason   string `json:"reason"`
}

type UserBanDto struct {
	UserBan
	Type string `json:"type"`
}

func (u *UserBan) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *UserBan) ToDto() []byte {
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
