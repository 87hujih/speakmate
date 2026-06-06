.PHONY: fmt fmt-check tidy tidy-check vet test build ci clean

APP_BIN ?= bin/speakmate
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

ci: fmt-check tidy-check vet test build

clean:
	rm -rf bin coverage.out
