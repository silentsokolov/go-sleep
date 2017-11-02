.PHONY: clean build test prepare release

VERSION = $(shell git describe --tags --always)
NAME := go-sleep

prepare:
	@go get github.com/jteeuwen/go-bindata/...
	@go get github.com/mitchellh/gox/...
	@echo Now you should be ready to run "make"

build: test
	@go build -ldflags "-X main.version=${VERSION}" -o go-sleep

clean:
	@rm -rf go-sleep-* go-sleep build templates.go

release: test
	gox \
	-ldflags="-X main.VERSION=${VERSION}" \
	-os="linux" \
	-output="build/{{.Dir}}_{{.OS}}_{{.Arch}}"

version:
	@echo $(VERSION)

test:
	@go generate
	@go test -race -v -parallel 4 ./...
