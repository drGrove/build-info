ARG BUILDER_IMAGE=golang
ARG BUILDER_SHA
ARG PKGVER
ARG EXPECTED_SHA
ARG SOURCE_DATE_EPOCH

FROM ${BUILDER_IMAGE}@sha256:${BUILDER_SHA} as builder
ARG PKGVER
ARG EXPECTED_SHA
ARG SOURCE_DATE_EPOCH

ADD go.mod /go/build-info/
ADD go.sum /go/build-info/
WORKDIR /go/build-info/
RUN go mod download
ADD . /go/build-info/
RUN make VERSION=${PKGVER} build-date=${SOURCE_DATE_EPOCH} build-info
# RUN echo "${EXPECTED_SHA} build-info" | sha512sum -c

FROM scratch
COPY --from=builder /go/build-info/build-info /build-info
