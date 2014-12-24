PREFIX ?= /usr/local
NAME=gounpack

all: deps test build

fmt:
	@find . -maxdepth 2 -name '*.go' -exec gofmt -w=true {} \;

test:
	@find . -maxdepth 2 -name '*_test.go' -printf "%h\n" | uniq | xargs go test

deps:
	go get -d -v

hack:
	@mkdir -p src/github.com/martinp
	@ln -sfn $(CURDIR) src/github.com/martinp/$(NAME)

install:
	cp -p bin/$(NAME) $(PREFIX)/bin/$(NAME)

build:
	@mkdir -p bin
	go build -o bin/$(NAME)

docker-image:
	docker build -t martinp/gounpack .
