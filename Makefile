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

yacc:
	goyacc -o transformers/basic/grammar.y.go transformers/basic/grammar.y
	goyacc -o transformers/advanced/grammar.y.go transformers/advanced/grammar.y

.PHONY: clean install test