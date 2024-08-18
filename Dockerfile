ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION} AS builder
COPY . /src
WORKDIR /src

ARG GOOS=linux


ENV CGO_ENABLED=0
ENV GOOS=${GOOS}
ENV GOARCH=${BUILDARCH}

RUN echo "GOARCH ENV: $BUILDARCH"
RUN GOARCH=${GOARCH} go build -o /src/bin/config-keeper-api
COPY ./migrations /src/bin/migrations

FROM alpine:latest
COPY --from=builder /src/bin /src/bin
WORKDIR /src/bin

ENV PRODUCTION=true
ENV PORT=8080
ENV TRACER_URL=http://localhost:14268/api/traces
ENV TRACER_NAME=test

ENTRYPOINT ["/src/bin/config-keeper-api"]


