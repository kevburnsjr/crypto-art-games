package controller

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

type privacyPolicy struct {
	log *logrus.Logger
}

func (c privacyPolicy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	t, err := template.ParseFiles("./template/privacyPolicy.html")
	if check(err, w, c.log) {
		return
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
