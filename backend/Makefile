#!/usr/bin/make -f
BIN=keg-scale

all: deps lint build test

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

test:
	go test -race ./...

build:
	go build -o $(BIN)

.PHONY: deps
deps:
	go mod tidy && go mod verify