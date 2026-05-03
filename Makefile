BIN     := marktex
CMD     := ./cmd/marktex
GOFLAGS := -trimpath

.PHONY: all build test lint clean install

all: build

## build: compile the marktex binary into ./bin/
build:
	@mkdir -p bin
	go build $(GOFLAGS) -o bin/$(BIN) $(CMD)

## test: run all tests
test:
	go test ./...

## test-verbose: run all tests with verbose output
test-verbose:
	go test -v ./...

## lint: run go vet (add golangci-lint if available)
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found; skipping extended lint"; \
	fi

## install: install marktex to GOPATH/bin
install:
	go install $(GOFLAGS) $(CMD)

## clean: remove build artifacts
clean:
	rm -rf bin/

## run: build and run with stdin (usage: echo "# hello" | make run)
run: build
	./bin/$(BIN)

## example: convert the sample file (if it exists)
example: build
	@if [ -f testdata/sample.md ]; then \
		./bin/$(BIN) -standalone testdata/sample.md; \
	else \
		echo "No testdata/sample.md found. Create one to test."; \
	fi

help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
