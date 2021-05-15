package entity

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/asaskevich/govalidator"
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

func UserFromHelix(u helix.User, secret string) User {
	normalized, err := govalidator.NormalizeEmail(u.Email)
	if err != nil {
		log.Println("Error normalizing email address:", u.Email)
		normalized = u.Email
	}
	hash := sha256.Sum256([]byte(secret + normalized))
	user := User(u)
	user.Email = base64.StdEncoding.EncodeToString(hash[:])
	return user
}
