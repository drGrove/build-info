VERSION ?= $(shell git describe --tags --always)
amd-builder-digest := d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e
arm-builder-digest := 533ea8ef7688aacf5626e18d71dd7ae6a4b02f8c226ae0ec461fd43acb815159
uname_m := $(shell uname -m)
uname_s := $(shell uname -s)
SOURCE_DATE_EPOCH ?= $(shell git log -1 --pretty=%ct)

ifeq ($(uname_m), x86_64)
armbuildflags := buildx build --platform=linux/arm64
CFLAGS ?= -march=x86-64 -mtune=generic -O2 -pipe -fno-plt -fexceptions \
            -Wp,-D_FORTIFY_SOURCE=2 -Wformat -Werror=format-security \
						-fstack-clash-protection -fcf-protection
CXX_FLAGS ?= $(CFLAGS) -Wp,-D_GLIBCXX_ASSERTIONS
LDFLAGS ?= -Wl,-O1,--sort-common,--as-needed,-z,relro,-z,now
else
CFLAGS ?= -march=armv8-a -O2 -pipe -fstack-protector-strong -fno-plt -fexceptions \
        -Wp,-D_FORTIFY_SOURCE=2 -Wformat -Werror=format-security \
				-fstack-clash-protection
CXXFLAGS ?= $(CFLAGS) -Wp,-D_GLIBCXX_ASSERTIONS
LDFLAGS ?= -Wl,-O1,--sort-common,--as-needed,-z,relro,-z,now
endif

ifeq ($(uname_s), Linux)
build-date := $(shell date -u '+%Y%m%d-%H:%M:%S' --date=@$(SOURCE_DATE_EPOCH))
else ifeq ($(uname_s), Darwin)
build-date := $(shell date -j -f '%s' $(SOURCE_DATE_EPOCH) '+%Y%m%d-%H:%M:%S')
endif


build-info:
	env CGO_LDFLAGS="$(LDFLAGS)" \
	CGO_CFLAGS="$(CFLAGS)" \
	CGO_CPPFLAGS="$(CPPFLAGS)" \
	CGO_CXXFLAGS="$(CXXFLAGS)" \
	go build \
		-trimpath \
		-buildmode=pie \
		-modcacherw \
		-buildvcs=false \
    -ldflags "-linkmode external -buildid= -extldflags \"$(LDFLAGS)\" \
		-X main.Version=$(VERSION) \
		-X main.BuildDate=$(build-date)" \
	./cmd/build-info

.PHONY: images
images:
	docker buildx build \
		--platform linux/arm64/v8,linux/amd64 \
		--build-arg SOURCE_DATE_EPOCH=$(build-date) \
		--build-arg PKGVER=$(VERSION) \
		--build-arg BUILDER_SHA=$(arm-builder-digest) \
		--build-arg EXPECTED_SHA=$(arm-expected-sha) \
		$(EXTRA_FLAGS) \
		-t drgrove/build-info:$(VERSION) \
		.

.PHONY: image
image: image-$(uname_m)

.PHONY: image-arm64
image-arm64:
	docker build \
		--build-arg SOURCE_DATE_EPOCH=$(build-date) \
		--build-arg PKGVER=$(VERSION) \
		--build-arg BUILDER_SHA=$(arm-builder-digest) \
		--build-arg EXPECTED_SHA=$(arm-expected-sha) \
		$(EXTRA_FLAGS) \
		-t drgrove/build-info:$(VERSION)-arm64 \
		.

.PHONY: image-x86_64
image-x86_64:
	docker build \
		--build-arg SOURCE_DATE_EPOCH=$(build-date) \
		--build-arg PKGVER=$(VERSION) \
		--build-arg BUILDER_SHA=$(amd-builder-digest) \
		--build-arg EXPECTED_SHA=$(amd-expected-sha) \
		$(EXTRA_FLAGS) \
		-t drgrove/build-info:$(VERSION)-amd64 \
		.
