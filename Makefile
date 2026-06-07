.PHONY: fmt fmt-check tidy tidy-check vet test build build-migrate migrate ci clean

APP_BIN ?= bin/speakmate
MIGRATE_BIN ?= bin/migrate
MIGRATIONS_DIR ?= migrations
MIGRATE_TIMEOUT ?= 60
GO_PACKAGES := ./...

fmt:
	gofmt -w cmd internal

fmt-check:
	@unformatted="$$(gofmt -l cmd internal)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Go files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

tidy:
	go mod tidy

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

vet:
	go vet $(GO_PACKAGES)

test:
	go test -race -coverprofile=coverage.out $(GO_PACKAGES)

build:
	go build -trimpath -o $(APP_BIN) ./cmd/server

build-migrate:
	go build -trimpath -o $(MIGRATE_BIN) ./cmd/migrate

migrate:
	go run ./cmd/migrate -dir $(MIGRATIONS_DIR) -timeout $(MIGRATE_TIMEOUT)

ci: fmt-check tidy-check vet test build build-migrate

clean:
	rm -rf bin coverage.out
