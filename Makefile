all: test build

build:
	manabuild mquery-sru

install:
	cp ./mquery-sru /usr/local/bin

clean:
	rm mquery-sru

test:
	go test ./...

rtest:
	go test -race ./...

.PHONY: clean install test