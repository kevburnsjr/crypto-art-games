VERSION := $(shell git describe --tags)
HASH := $(shell git rev-parse --short HEAD)
# PROJECTNAME := $(shell basename "$(PWD)")
PROJECTNAME := "crypto-art-games"

LDFLAGS := -ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.Hash=$(HASH)'"

clean:
	rm -rf ./dist/*

build-stage:
	go build $(LDFLAGS) -o dist/$(PROJECTNAME) main.go

build-prod:
	go build $(LDFLAGS) -o dist/$(PROJECTNAME) main.go

init-stage:
	sudo cp ./deploy/systemd/$(PROJECTNAME).service /etc/systemd/system/$(PROJECTNAME).service
	sudo systemctl enable $(PROJECTNAME)
	sudo systemctl start $(PROJECTNAME)
