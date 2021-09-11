XGOARCH := amd64
XGOOS := linux
XBIN := $(XGOOS)_$(XGOARCH)/unp

all: test vet checkfmt install

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

install:
	go install ./...

checkfmt:
	@bash -c "diff --line-format='%L' <(echo -n) <(gofmt -l .)"

xinstall:
	env GOOS=$(XGOOS) GOARCH=$(XGOARCH) go install ./...

publish:
ifndef DEST_PATH
	$(error DEST_PATH must be set when publishing)
endif
	rsync -a $(GOPATH)/bin/$(XBIN) $(DEST_PATH)/$(XBIN)
	@sha256sum $(GOPATH)/bin/$(XBIN)
