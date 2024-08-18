ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION} AS builder
COPY . /src
WORKDIR /src

ARG GOOS=linux

RUN ARCH=$(uname -m) && \
    case $ARCH in \
        x86_64) GOARCH=amd64 ;; \
        aarch64) GOARCH=arm64 ;; \
        arm64) GOARCH=arm64 ;; \
        armv7l) GOARCH=arm ;; \
        i386|i686) GOARCH=386 ;; \
        s390x) GOARCH=s390x ;; \
        *) GOARCH=amd64 ;; \
    esac && \
    echo "GOARCH: ${GOARCH}" && \
    export GOARCH=${GOARCH}

ENV CGO_ENABLED=0
ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}

RUN echo "GOARCH ENV: $GOARCH"
RUN go build -o /src/bin/config-keeper-api
COPY ./migrations /src/bin/migrations

FROM alpine:latest
COPY --from=builder /src/bin /src/bin
WORKDIR /src/bin

ENV PRODUCTION=true
ENV PORT=8080
ENV TRACER_URL=http://localhost:14268/api/traces
ENV TRACER_NAME=test

ENTRYPOINT ["/src/bin/config-keeper-api"]


