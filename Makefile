all: build run

.PHONY: run
run: build
	./searchinform

searchinform: $(shell find $(pwd) -name '*.go' | tr '\n' ' ')
	go build .

.PHONY: build
build: searchinform

.PHONY: tests
tests:
	go test -cover -v github.com/searchinform/cache \
						github.com/searchinform/provider
