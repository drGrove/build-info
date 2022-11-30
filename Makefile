git-version := $(shell git describe --tags --always)

.PHONY: build-info
build-info:
	go build \
		-trimpath \
		-buildmode=pie \
		-modcacherw \
		-ldflags "-linkmode external ${LDFLAGS} -X github.com/drGrove/build-info/cmd/build-info.Version=$(git-version)" \
		./cmd/build-info
