package entity

import (
	"encoding/json"
	"fmt"
	"time"

	// "github.com/asaskevich/govalidator"
	"github.com/nicklaw5/helix"
)

type User struct {
	helix.User
	UserID     uint32                 `json:"userID"`
	Policy     bool                   `json:"policy"`
	Timeout    uint32                 `json:"timeout"`
	Banned     bool                   `json:"banned"`
	Mod        bool                   `json:"mod"`
	Buckets    map[uint16]*UserBucket `json:"buckets"`
	Created    uint32                 `json:"created"`
}

type UserDto struct {
	Type        string `json:"type"`
	ID          uint32 `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

func (u *User) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *User) ToDto(userID uint32) []byte {
	b, _ := json.Marshal(UserDto{
		"new-user",
		userID,
		u.Login,
		u.DisplayName,
	})
	return b
}

func (u *User) GetBucket(boardID uint16) *UserBucket {
	if u == nil {
		return nil
	}
	if u.Buckets == nil {
		u.Buckets = map[uint16]*UserBucket{}
	}
	if _, ok := u.Buckets[boardID]; !ok {
		u.Buckets[boardID] = NewUserBucket(time.Now())
	}
	return u.Buckets[boardID]
}

func (u *User) IDHex() string {
	return fmt.Sprintf("%06x", u.UserID)
}

func (u *User) Active(t time.Time) error {
	var to = time.Unix(int64(u.Timeout), 0)
	if u.Timeout > 0 && t.Before(to) {
		return fmt.Errorf("Timed out. (%v) remaining", to.Sub(t).Truncate(time.Second))
	}
	return nil
}

func UserFromJson(b []byte) *User {
	var u User
	err := json.Unmarshal(b, &u)
	if err != nil {
		return nil
	}
	return &u
}

func UserFromHelix(u helix.User, secret string) *User {
	/*
		normalized, err := govalidator.NormalizeEmail(u.Email)
		if err != nil {
			log.Println("Error normalizing email address:", u.Email)
			normalized = u.Email
		}
		u.Email = normalized
	*/
	return &User{User: u}
}
