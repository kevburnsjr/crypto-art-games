package main

import (
	"flag"
	"log"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/kevburnsjr/crypto-art-games/internal"
	"github.com/kevburnsjr/crypto-art-games/internal/config"
)

var config_path *string = flag.String("conf", "config.yml", "Location of config file")

func main() {
	flag.Parse()
	var yamlFile, err = ioutil.ReadFile(*config_path)
	if err != nil {
		log.Fatalf("Config file not found - %s", *config_path)
	}
	var cfg = config.Api{}
	yaml.Unmarshal(yamlFile, &cfg)

	var app = internal.NewApi(&cfg)

	log.Println("Starting api")
	app.Start()

	var stop = make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Stopping api")
	app.Stop(10 * time.Second)
}
