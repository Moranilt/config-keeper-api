PRODUCTION=false
PORT=8080

VAULT_TOKEN=test-token
VAULT_HOST=http://localhost:8200
VAULT_MOUNT_PATH=secret
VAULT_DB_CREDS_PATH=test/db

TRACER_URL=http://localhost:14268/api/traces
TRACER_NAME=test

DEFAULT_ENV=PRODUCTION=$(PRODUCTION) PORT=$(PORT)

VAULT_ENV = VAULT_TOKEN=$(VAULT_TOKEN) VAULT_HOST=$(VAULT_HOST) VAULT_MOUNT_PATH=$(VAULT_MOUNT_PATH) VAULT_DB_CREDS_PATH=$(VAULT_DB_CREDS_PATH)

TRACER_ENV = TRACER_URL=$(TRACER_URL) TRACER_NAME=$(TRACER_NAME)

ENVIRONMENT = $(DEFAULT_ENV) $(VAULT_ENV) $(TRACER_ENV)

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