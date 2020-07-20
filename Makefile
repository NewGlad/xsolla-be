.PHONY: build

build:
	go build -v ./cmd/newsapi/

run:
	go run ./cmd/newsapi	

.DEFAULT_GOAL := run