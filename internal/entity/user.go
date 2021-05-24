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
	UserID  uint16      `json:"userID"`
	Policy  bool        `json:"policy"`
	Timeout time.Time   `json:"timeout"`
	Banned  bool        `json:"banned"`
	Mod     bool        `json:"mod"`
	Bucket  *UserBucket `json:"bucket"`
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

type UserBucket struct {
	Size      uint8  `json:"size"`
	Rate      uint8  `json:"rate"`
	Level     uint8  `json:"level"`
	Timestamp uint32 `json:"timestamp"`
}

func NewUserBucket() *UserBucket {
	return &UserBucket{Size: 8, Level: 32, Rate: 4, Timestamp: uint32(time.Now().Unix())}
}

func (b *UserBucket) AdjustLevel() {
	var delta = time.Now().Sub(time.Unix(int64(b.Timestamp), 0))
	var levelDelta = int(delta/time.Second) / int(60 / b.Rate)
	if levelDelta + int(b.Level) > int(b.Size * 4) {
		b.Level = b.Size * 4
	} else {
		b.Level += uint8(levelDelta)
	}
	b.Timestamp += uint32(levelDelta * int(60 / b.Rate))
}

func (b *UserBucket) Consume(n uint8) bool {
	b.AdjustLevel()
	if b.Level < n*4 {
		return false
	}
	b.Level -= n * 4
	return true
}

func (b *UserBucket) Credit(n uint8) {
	b.AdjustLevel()
	b.Level = b.Level + n*4
	if b.Level > b.Size*4 {
		b.Level = b.Size * 4
	}
}
