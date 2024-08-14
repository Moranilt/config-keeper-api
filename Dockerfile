FROM golang:1.21 AS builder

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/bin/test_project

COPY ./migrations /src/bin/migrations

FROM alpine:latest

ARG PRODUCTION
ENV PRODUCTION=$PRODUCTION

ARG PORT
ENV PORT=$PORT

ARG TRACER_URL
ENV TRACER_URL=$TRACER_URL

ARG TRACER_NAME
ENV TRACER_NAME=$TRACER_NAME

COPY --from=builder /src/bin /src/bin
WORKDIR /src/bin

EXPOSE $PORT
ENTRYPOINT ["/src/bin/test_project"]


