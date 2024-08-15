PRODUCTION=false
PORT=8080

DB_NAME=config-keeper
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=123456
DB_SSL_MODE=disable

TRACER_URL=http://localhost:14268/api/traces
TRACER_NAME=test

DEFAULT_ENV=PRODUCTION=$(PRODUCTION) PORT=$(PORT)

DB_ENV=DB_NAME=$(DB_NAME) DB_HOST=$(DB_HOST) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_SSL_MODE=$(DB_SSL_MODE)

TRACER_ENV = TRACER_URL=$(TRACER_URL) TRACER_NAME=$(TRACER_NAME)

ENVIRONMENT = $(DEFAULT_ENV) $(DB_ENV) $(TRACER_ENV)

run:
	$(ENVIRONMENT) go run .

run-race:
	$(ENVIRONMENT) go run -race .

test:
	go test ./...

test-cover:
	go test ./... -coverprofile=./cover

cover-html:
	go tool cover -html=./cover

docker-up:
	docker compose up -d

docker-down:
	docker compose down