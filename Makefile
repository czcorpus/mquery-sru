all: build-and-test

build:
	go generate ./cmd/service/fcs.go
	manabuild -cmd-dir service

build-and-test:
	manabuild -test -cmd-dir service mquery-sru

tools:
	go install github.com/mna/pigeon
	go install github.com/czcorpus/manabuild@latest

generate:
	go generate ./cmd/service/fcs.go

install:
	cp ./mquery-sru /usr/local/bin

clean:
	rm mquery-sru

test:
	go test $(go list ./... | grep -v "github.com/czcorpus/mquery-sru/cmd/testing")

itest:
	go test -v ./cmd/testing --args http://localhost:8989/

rtest:
	go test -race ./...

.PHONY: clean install test itest tools