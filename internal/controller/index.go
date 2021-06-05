package controller

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"encoding/csv"
	"log"
	"io"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"

	"github.com/kevburnsjr/crypto-art-games/internal/config"
	"github.com/kevburnsjr/crypto-art-games/internal/repo"
	sock "github.com/kevburnsjr/crypto-art-games/internal/socket"
)

var allJS = []string{
	"/js/lib/helpers.js",
	"/js/lib/timeago.min.js",
	"/js/lib/jsuri-1.1.1.js",
	"/js/lib/base64.js",
	"/js/lib/bitset.min.js",
	"/js/lib/nearestColor.js",
	"/js/lib/localforage.min.js",
	"/js/lib/polyfills.js",
	"/js/global.js",
	"/js/game.js",
	"/js/lib/object.js",
	"/js/lib/event.js",
	"/js/lib/socket.js",
	"/js/lib/dom.js",
	"/js/series.js",
	"/js/user.js",
	"/js/nav.js",
	"/js/board.js",
	"/js/palette.js",
	"/js/tile.js",
	"/js/frame.js",
}

type index struct {
	*oauth
	cfg      *config.Api
	log      *logrus.Logger
	hub      sock.Hub
	repoUser repo.User
}

var localizationMaps map[language.Base]localizationMap

type localizationMap map[string]string

var langMatcher = language.NewMatcher([]language.Tag{
	language.English,
	language.AmericanEnglish,
	language.BritishEnglish,
	language.Spanish,
	language.EuropeanSpanish,
	language.LatinAmericanSpanish,
})

func (l localizationMap) String(key string) string {
	return l[key]
}

func (c index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if localizationMaps == nil {
		localizationMaps = loadLocalization()
	}
	stdHeaders(w)
	if r.URL.Path == "/" {
		w.Header().Set("Location", "/pixel-compactor")
		w.WriteHeader(302)
		return
	}
	var mod bool
	user, _ := c.oauth.getUser(r, w)
	if user != nil {
		mod = user.Mod
	}
	var js = allJS
	if c.cfg.Minify {
		js = []string{"/js/min.js?v=" + c.cfg.Hash}
	}
	if c.cfg.Test {
		js = append(js, "/js/test.js")
	}

	locaMap := localizationMaps[getLang(r)]

	b := bytes.NewBuffer(nil)
	var indexTpl = template.Must(template.ParseFiles("./template/index.html"))
	err := indexTpl.Execute(b, struct {
		HOST string
		JS   []string
		Mod  bool
		Loca localizationMap
	}{
		c.cfg.Http.Host,
		js,
		mod,
		locaMap,
	})
	if check(err, w, c.log) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(b.Bytes())
}

func getLang(r *http.Request) language.Base {
	langTags, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	langTag, _, _ := langMatcher.Match(langTags...)
	base, _ := langTag.Base()
	return base
}

func check(err error, w http.ResponseWriter, log *logrus.Logger) bool {
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), 500)
		return true
	}
	return false
}

/* example {
 *     "en": {
 *         "Hello": "Hello",
 *         "World": "World"
 *     },
 *     "es": {
 *         "Hello": "Hola",
 *         "World": "Mundo"
 *     }
 * }
 */
func loadLocalization() map[language.Base]localizationMap {
    f, err := os.Open("i18n/strings.csv")
    if err != nil {
		log.Fatalln("Couldn't open localization csv file", err)
    }
    defer f.Close()

    csvr := csv.NewReader(f)
	langStrs, err := csvr.Read()
	if err != nil {
		log.Fatalln("Couldn't read localization csv file header", err)
	}
	langs := make([]language.Base, len(langStrs))
	m := map[language.Base]localizationMap{}
	for i, langStr := range langStrs {
		langs[i], _ = language.Make(langStr).Base()
		m[langs[i]] = localizationMap{}
	}
    for {
        row, err := csvr.Read()
        if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("Couldn't read csv", err)
        }
		for i, str := range row {
			m[langs[i]][row[0]] = str
		}
	}
	return m
}