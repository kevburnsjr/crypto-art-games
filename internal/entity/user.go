package entity

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	// "strconv"

	"github.com/asaskevich/govalidator"
	"github.com/nicklaw5/helix"
)

type User struct {
	helix.User
	Policy bool `json:"policy"`
}

type UserDto struct {
	Type string `json:"type"`
	ID uint16 `json:"id"`
	Login string `json:"login"`
	DisplayName string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

func (u *User) ToJson() []byte {
	b, _ := json.Marshal(u)
	log.Println(string(b))
	return b
}

func (u *User) ToDto(userID uint16) []byte {
	b, _ := json.Marshal(UserDto{
		"new-user",
		userID,
		u.Login,
		u.DisplayName,
		u.ProfileImageURL,
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

func UserFromHelix(u helix.User, secret string) User {
	normalized, err := govalidator.NormalizeEmail(u.Email)
	if err != nil {
		log.Println("Error normalizing email address:", u.Email)
		normalized = u.Email
	}
	hash := sha256.Sum256([]byte(secret + normalized))
	user := User{User: u}
	user.Email = base64.StdEncoding.EncodeToString(hash[:])
	return user
}
