.PHONY: test build run release lint

test:
	go test -v ./...

build:
	go build -o chinchon ./...

run:
	./chinchon

release:
	rm -rf dist && goreleaser

lint:
	golangci-lint run