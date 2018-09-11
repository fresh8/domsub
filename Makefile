SHELL := /bin/bash
PUB_SRC := cmd/clipublisher/*.go
SUB_SRC := cmd/clisubscriber/*.go

.PHONY: all
all: subscriber publisher

subscriber: $(SUB_SRC)
	go build -v -o subscriber ./cmd/clisubscriber

publisher: $(PUB_SRC)
	go build -v -o publisher ./cmd/clipublisher
