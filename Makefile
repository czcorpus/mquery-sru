all: build-and-test

build:
	manabuild mquery-sru

build-and-test:
	manabuild -test mquery-sru

tools:
	go install github.com/mna/pigeon
	go install github.com/czcorpus/manabuild

install:
	cp ./mquery-sru /usr/local/bin

clean:
	rm mquery-sru

test:
	go test ./...

rtest:
	go test -race ./...

.PHONY: clean install test tools