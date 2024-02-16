all: build-and-test

build:
	manabuild -cmd-dir service

build-and-test:
	manabuild -test -cmd-dir service mquery-sru

tools:
	go install github.com/mna/pigeon
	go install github.com/czcorpus/manabuild@latest

install:
	cp ./mquery-sru /usr/local/bin

clean:
	rm mquery-sru

test:
	go test ./...

rtest:
	go test -race ./...

.PHONY: clean install test tools