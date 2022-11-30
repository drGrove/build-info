version := $(shell git describe --tags --always)
branch := $(shell git symbolic-ref --short -q HEAD)
build-user := someone@builder
build-date := $(shell date -u '+%Y%m%d-%H:%M:%S' --date=$${SOURCE_DATE_EPOCH})
builder-digest := d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e
expected-sha := 6ccf81a481ec9f285bd5ac1334221ab146b21ee46e6cd4d3a200c8d1a5c640522dec8ba0617d75989625acebc50b90a2457640fefd618c43f1e8a008164db44f

.PHONY: build-info
build-info:
	go build \
		-trimpath \
		-buildmode=pie \
		-modcacherw \
		-ldflags "-linkmode external \
		-X main.Version=$(version) \
		-X main.Revision=$(version) \
		-X main.Branch=$(git-branch) \
		-X main.BuildUser=$(build-user) \
		-X main.BuildDate=$(build-date) \
		$${LDFLAGS}" \
		./cmd/build-info

.PHONY: image
image:
	docker build \
		--build-arg SOURCE_DATE_EPOCH=$(build-date) \
		--build-arg PKGVER=$(version) \
		--build-arg BUILDER_IMAGE=golang \
		--build-arg BUILDER_SHA=$(builder-digest) \
		--build-arg EXPECTED_SHA=$(expected-sha) \
		-t drgrove/build-info:$(version) \
		.
