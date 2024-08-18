FROM golang:1.23 AS builder
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/bin/config-keeper-api
COPY ./migrations /src/bin/migrations

FROM alpine:latest
COPY --from=builder /src/bin /src/bin
WORKDIR /src/bin

ENTRYPOINT ["/src/bin/config-keeper-api"]


