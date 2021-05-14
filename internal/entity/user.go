package entity

import (
	"encoding/json"

	"github.com/nicklaw5/helix"
)

type User helix.User

func (u *User) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func UserFromJson(b []byte) *User {
	var u User
	err := json.Unmarshal(b, &u)
	if err != nil {
		return nil
	}
	return &u
}
