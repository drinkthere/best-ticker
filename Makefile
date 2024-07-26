# strategy version
strategy ?= 0.1.0

# Repo info
GIT_COMMIT ?= git-$(shell git rev-parse --short HEAD)
TARGETS := darwin/amd64 linux/amd64 windows/amd64

$(info $(GIT_COMMIT) )

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

$(info $(GOBIN))

build: fmt vet
	go mod vendor
	go build -o bin/bt ./*.go

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: golangci
	gofmt -s -w client config container context message polling trading utils watchdog
	$(GOLANGCILINT) run --timeout 10m -E errcheck,gosimple,goimports  ./...

.PHONY: proto
proto:
	protoc --go_out=. ./protocol/proto/okxticker.proto
	protoc --go_out=. ./protocol/proto/okxorderbook.proto

golangci:
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	echo 'installing golangci-lint-$(GOLANGCILINT_VERSION)' ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCILINT_VERSION) ;\
	echo 'Install succeed' ;\
	}
GOLANGCILINT=$(GOBIN)/golangci-lint
else
GOLANGCILINT=$(shell which golangci-lint)
endif
