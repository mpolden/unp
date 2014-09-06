all: test

fmt:
	gofmt -w=true *.go

test:
	go test
