.PHONY: all build install test vet fmt help

all: test build ## test then build

build: ## build the asl binary
	go build -o asl .

install: ## install asl into GOPATH/bin
	go install .

test: ## run analyzer tests
	go test ./...

vet: ## run go vet
	go vet ./...

fmt: ## format sources (testdata excluded)
	gofumpt -w main.go constraintname funcpartition functypedup internal methodinterleave methodpartition testorder topdecl typeblockgap
	goimports -local github.com/CMGS/asl -w main.go constraintname funcpartition functypedup internal methodinterleave methodpartition testorder topdecl typeblockgap

help: ## show this help
	@grep -E '^[a-z]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-10s %s\n", $$1, $$2}'
