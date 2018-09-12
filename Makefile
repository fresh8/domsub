SHELL := /bin/bash
PUB_SRC := cmd/clipublisher/*.go
SUB_SRC := cmd/clisubscriber/*.go

.PHONY: all
all: subscriber publisher

subscriber: $(SUB_SRC)
	go build -v -o subscriber ./cmd/clisubscriber

publisher: $(PUB_SRC)
	go build -v -o publisher ./cmd/clipublisher

.PHONY: drain
drain: subscriber
	./subscriber -stop 10 # drain the queue

.PHONY: s1
s1: subscriber publisher drain
	./publisher # publish a message
	./subscriber -start 0 -stop 60 -ack 30 > sub1.log & # subscriber 1 should receive and ack published message
	./subscriber -start 10 -stop 50 -ack 0 > sub2.log # subscriber 2 should see nothing
	cat sub1.log sub2.log

.PHONY: s2
s2: subscriber publisher drain
	./publisher # publish a message
	./subscriber -start 0 -stop 10 -ack 30 -fail true > sub1.log & # subscriber 1 should receive and die
	./subscriber -start 10 -stop 60 -ack 0 > sub2.log # subscriber 2 should see message and ack after ack deadline has passed
	cat sub1.log sub2.log

