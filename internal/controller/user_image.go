package controller

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/kevburnsjr/crypto-art-games/internal/repo"
)

func newUserImage(rUser repo.User) *userImage {
	return &userImage{rUser}
}

type userImage struct {
	repoUser repo.User
}

func (c userImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var location string
	user, err := c.repoUser.FindByUserID(uint16(userID))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		println(err.Error())
		return
	}
	if user == nil {
		location = "/i/q.png"
	} else {
		location = user.ProfileImageURL
	}
	w.Header().Set("Location", location)
	w.Header().Set("Cache-Control", "max-age=900, stale-while-revalidate=86400")
	w.WriteHeader(302)
	return
}
