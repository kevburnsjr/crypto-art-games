package entity

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nicklaw5/helix"
)

type User struct {
	helix.User
	UserID  uint16    `json:"userID"`
	Policy  bool      `json:"policy"`
	Timeout time.Time `json:"timeout"`
	Banned  bool      `json:"banned"`
	Mod     bool      `json:"mod"`
}

type UserDto struct {
	Type            string `json:"type"`
	ID              uint16 `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
}

func (u *User) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *User) ToDto(userID uint16) []byte {
	b, _ := json.Marshal(UserDto{
		"new-user",
		userID,
		u.Login,
		u.DisplayName,
		u.ProfileImageURL,
		u.OfflineImageURL,
	})
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
