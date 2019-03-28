GO_EXECUTABLE ?= go
DIST_DIRS := find * -type d -exec
VERSION ?= $(shell git describe --tags)

build:
	${GO_EXECUTABLE} build -o sbucket -ldflags "-X main.version=${VERSION}" sbucket.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./sbucket ${DESTDIR}/usr/local/bin/sbucket

.PHONY: build install
