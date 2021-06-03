package entity

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nicklaw5/helix"
)

type User struct {
	helix.User
	UserID  uint32                 `json:"userID"`
	Policy  bool                   `json:"policy"`
	Timeout time.Time              `json:"timeout"`
	Banned  bool                   `json:"banned"`
	Mod     bool                   `json:"mod"`
	Buckets map[uint16]*UserBucket `json:"buckets"`
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

func UserFromJson(b []byte) *User {
	var u User
	err := json.Unmarshal(b, &u)
	if err != nil {
		return nil
	}
	return &u
}

func UserFromHelix(u helix.User, secret string) *User {
	normalized, err := govalidator.NormalizeEmail(u.Email)
	if err != nil {
		log.Println("Error normalizing email address:", u.Email)
		normalized = u.Email
	}
	hash := sha256.Sum256([]byte(secret + normalized))
	user := &User{User: u}
	user.Email = base64.StdEncoding.EncodeToString(hash[:])
	return user
}
