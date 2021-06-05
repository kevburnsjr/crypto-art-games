package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

type termsOfService struct {
	log *logrus.Logger
}

func (c termsOfService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	t, err := template.ParseFiles("./template/termsOfService.html")
	if check(err, w, c.log) {
		return
	}
	if getLang(r).String() == "es" {
		t, err = template.ParseFiles("./template/termsOfService.es.html")
		if check(err, w, c.log) {
			return
		}
	}
	b := bytes.NewBuffer(nil)
	err = t.Execute(b, struct{}{})
	if check(err, w, c.log) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}
