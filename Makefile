.PHONY: gru minion clean 

# Registry used for publishing images
#REGISTRY?=${REGISTRY_PREFIX}overpass

# Default tag and architecture. Can be overridden
TAG?=$(shell git describe --tags --dirty --always)
ifeq ($(TAG),)
	TAG=latest
endif

ifeq ($(findstring dirty,$(TAG)), dirty)
	TAG=latest
endif


#gru: $(shell find . -type f  -name '*.go')
gru: 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w -X github.com/ski2per/gru/gru.Version=$(TAG) -extldflags "-static"' \
	-o dist/gru bin/gru/gru.go 

minion: 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w -X github.com/ski2per/gru/gru.Version=$(TAG) -extldflags "-static"' \
	-o dist/minion bin/minion/minion.go 

build: gru minion

clean:
	rm -f dist/*
