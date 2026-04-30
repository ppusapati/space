SHELL := /usr/bin/env bash
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

SERVICES := earthobs satsubsys groundstation geoint

.PHONY: tools proto sqlc generate build test vet tidy clean

tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.69.0
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
	go install ariga.io/atlas/cmd/atlas@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

proto:
	buf lint
	buf generate

sqlc:
	@for s in $(SERVICES); do \
		echo ">> sqlc generate for $$s"; \
		(cd services/$$s && sqlc generate) || exit 1; \
	done

generate: proto sqlc

tidy:
	cd pkg && go mod tidy
	@for s in $(SERVICES); do (cd services/$$s && go mod tidy) || exit 1; done

vet:
	cd pkg && go vet ./...
	@for s in $(SERVICES); do (cd services/$$s && go vet ./...) || exit 1; done

build:
	@for s in $(SERVICES); do \
		echo ">> build $$s"; \
		(cd services/$$s && CGO_ENABLED=0 go build -o bin/$$s ./cmd/$$s) || exit 1; \
	done

test:
	cd pkg && go test ./...
	@for s in $(SERVICES); do (cd services/$$s && go test ./...) || exit 1; done

clean:
	rm -rf services/*/bin
