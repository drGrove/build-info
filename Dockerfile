ARG BUILDER_IMAGE
ARG BUILDER_SHA
ARG PKGVER
ARG EXPECTED_SHA
ARG SOURCE_DATE_EPOCH

FROM ${BUILDER_IMAGE}@sha256:${BUILDER_SHA} as builder
ARG PKGVER
ARG EXPECTED_SHA
ARG SOURCE_DATE_EPOCH

# Config used by Arch Linux to get reproducible builds
ENV CFLAGS="-march=x86-64 -mtune=generic -O2 -pipe -fno-plt -fexceptions \
            -Wp,-D_FORTIFY_SOURCE=2 -Wformat -Werror=format-security \
            -fstack-clash-protection -fcf-protection"
ENV CXX_FLAGS="$CFLAGS -Wp, -D_GLIBCXX_ASSERTIONS"
ENV LDFLAGS="-Wl,-O1,--sort-common,--as-needed,-z,relro,-z,now"
ENV LTOFLAGS="-flto=auto"
ENV CGO_LDFLAGS="${LDFLAGS}"
ENV CGO_CFLAGS="${CFLAGS}"
ENV GCO_CPPFLAGS="${CPPFLAGS}"
ENV GCO_CXXFLAGS="${CXXFLAGS}"

ADD . /go/build-info/
WORKDIR /go/build-info/
RUN go build \
    -trimpath \
    -buildmode=pie \
    -modcacherw \
    -ldflags "-linkmode external -extldflags ${LDFLAGS} \
    -X main.Version=${PKGVER} \
    -X main.Revision=${PKGVER} \
    -X main.Branch=source \
    -X main.BuildUser=someone@builder \
    -X main.BuildDate=${SOURCE_DATE_EPOCH}" \
  ./cmd/build-info

# RUN echo "${EXPECTED_SHA} build-info" | sha512sum -c

FROM scratch
COPY --from=builder /go/build-info/build-info /build-info
