package controller

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tdewolff/minify"
	jsmin "github.com/tdewolff/minify/js"
)

type staticMinJS struct {
	prefix string
	hash string
}

func (c staticMinJS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stdHeaders(w)
	// 304
	etag := r.Header.Get("If-None-Match")
	w.Header().Set("Etag", c.hash)
	w.Header().Set("Cache-Control", "max-age=2592000")
	if len(c.hash) > 0 && etag == c.hash {
		w.WriteHeader(304)
		return
	}
	mediatype := "text/javascript"
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	m := minify.New()
	m.AddFunc(mediatype, jsmin.Minify)
	for _, path := range allJS {
		src, err := ioutil.ReadFile(c.prefix + path)
		if err != nil {
			log.Println(err)
			continue;
		}
		min, _ := m.Bytes(mediatype, src)
		w.Write(min)
	}
}
