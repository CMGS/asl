.PHONY: all build install test vet lint fmt fmt-check deps coverage clean help golangci-lint gofumpt goimports

GOOSES ?= linux darwin
GOFILES := $(shell find . -name '*.go' -not -path '*/testdata/*' -not -path './bin/*')

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOLANGCILINT_VERSION ?= v2.12.2
GOLANGCILINT_ROOT := $(LOCALBIN)/golangci-lint-$(GOLANGCILINT_VERSION)
GOLANGCILINT := $(GOLANGCILINT_ROOT)/golangci-lint

GOFUMPT_VERSION ?= v0.10.0
GOIMPORTS_VERSION ?= v0.48.0
GOFMT := $(LOCALBIN)/gofumpt-$(GOFUMPT_VERSION)
GOIMPORTS := $(LOCALBIN)/goimports-$(GOIMPORTS_VERSION)

golangci-lint: $(GOLANGCILINT)
$(GOLANGCILINT):
	GOBIN=$(GOLANGCILINT_ROOT) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCILINT_VERSION)

gofumpt: $(GOFMT)
$(GOFMT): | $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	mv $(LOCALBIN)/gofumpt $(GOFMT)

goimports: $(GOIMPORTS)
$(GOIMPORTS): | $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	mv $(LOCALBIN)/goimports $(GOIMPORTS)

all: deps fmt lint test build ## Full pipeline: deps, fmt, lint, test, build

deps: ## Tidy Go modules
	go mod tidy

build: ## Build the asl binary
	CGO_ENABLED=0 go build -o asl .

install: ## Install asl into GOPATH/bin
	go install .

test: vet ## Run analyzer tests with race detection and coverage
	go test -race -timeout 120s -count=1 -cover -coverprofile=coverage.out ./...

coverage: test ## Generate and display coverage report
	go tool cover -func=coverage.out

vet: ## Run go vet on every target OS
	@for goos in $(GOOSES); do \
		echo "==> go vet GOOS=$$goos"; \
		GOOS=$$goos go vet ./... || exit 1; \
	done

lint: golangci-lint ## Run golangci-lint on every target OS
	@for goos in $(GOOSES); do \
		echo "==> golangci-lint GOOS=$$goos"; \
		GOOS=$$goos $(GOLANGCILINT) run ./... || exit 1; \
	done

fmt: gofumpt goimports ## Format sources with gofumpt and goimports (testdata excluded)
	$(GOFMT) -extra -l -w $(GOFILES)
	$(GOIMPORTS) -l -w --local github.com/CMGS/asl $(GOFILES)

fmt-check: gofumpt goimports ## Check formatting (fails if files need formatting)
	@test -z "$$($(GOFMT) -extra -l $(GOFILES))" || { echo "Files need formatting (gofumpt):"; $(GOFMT) -extra -l $(GOFILES); exit 1; }
	@test -z "$$($(GOIMPORTS) -l --local github.com/CMGS/asl $(GOFILES))" || { echo "Files need formatting (goimports):"; $(GOIMPORTS) -l --local github.com/CMGS/asl $(GOFILES); exit 1; }

clean: ## Remove build artifacts, coverage files, and test cache
	rm -f asl asl-linux-*
	rm -rf bin/
	rm -f coverage.out
	go clean -testcache

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
