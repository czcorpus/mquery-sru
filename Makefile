all: test build

build:
	manabuild fcs

install:
	cp ./fcs /usr/local/bin

clean:
	rm fcs

test:
	go test ./...

rtest:
	go test -race ./...

.PHONY: clean install test