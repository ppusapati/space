SHELL := /usr/bin/env bash
GOBIN := $(shell go env GOPATH)/bin
CARGO := cargo
export PATH := $(GOBIN):$(PATH)

# Each service folder under services/. Populated incrementally as Go
# services are implemented.
SERVICES :=

# Rust workspaces. `compute` hosts the host-side data/RF/imagery crates;
# `flight` hosts the embedded-target ADCS / C&DH / EPS crates that will
# eventually be cross-compiled for satellite hardware.
RUST_WORKSPACES := compute flight

# Python ML packages and worker daemons.
PY_PKGS := ml/packages/ml_serving ml/packages/eo_ml ml/packages/gi_ml \
           ml/workers/eo-ml-worker ml/workers/gi-ml-worker

.PHONY: tools proto sqlc generate build test vet tidy clean rust-test rust-clippy py-test

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
	@if [ -d pkg ]; then cd pkg && go test ./...; fi
	@for s in $(SERVICES); do (cd services/$$s && go test ./...) || exit 1; done

rust-test:
	@for w in $(RUST_WORKSPACES); do \
		echo ">> cargo test --workspace ($$w)"; \
		(cd $$w && $(CARGO) test --workspace --all-targets) || exit 1; \
	done

rust-clippy:
	@for w in $(RUST_WORKSPACES); do \
		echo ">> cargo clippy --workspace ($$w)"; \
		(cd $$w && $(CARGO) clippy --workspace --all-targets --all-features -- -D warnings) || exit 1; \
	done

py-test:
	@for p in $(PY_PKGS); do \
		echo ">> pytest $$p"; \
		(cd $$p && python3 -m pytest tests/ -q) || exit 1; \
	done

clean:
	rm -rf services/*/bin
	@for w in $(RUST_WORKSPACES); do (cd $$w && $(CARGO) clean) || true; done
