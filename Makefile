PREFIX ?= /usr/local
NAME=gounpack

all: test build

fmt:
	gofmt -w=true *.go

test:
	go test

install:
	cp -p bin/$(NAME) $(PREFIX)/bin/$(NAME)

build:
	@mkdir -p bin
	go build -o bin/$(NAME)
