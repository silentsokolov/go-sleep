.PHONY: clean build

GO ?= go
VERSION = $(shell git describe --tags --always)
NAME := go-sleep

build: get
	${GO} build -ldflags "-X main.version=${VERSION}" -o go-sleep

clean:
	@rm -rf go-sleep-* go-sleep build debian

get:
	${GO} get -t -v ./...

install-gox:
	${GO} get github.com/mitchellh/gox

release: clean install-gox
	gox \
	-ldflags="-X main.VERSION=${VERSION}" \
	-os="linux" \
	-output="build/{{.Dir}}_{{.OS}}_{{.Arch}}"

	cp config.sample.toml build/
	cp -R templates build/
	cd build; \
	for target in go-sleep_*; \
	do \
		cp $$target $(NAME); \
		tar -cvzf $$target.tar.gz $(NAME) config.sample.toml templates; \
		rm $(NAME); \
		rm $$target; \
	done; \
	rm -R templates; \
	rm config.sample.toml;

version:
	@echo $(VERSION)
