SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

API_ROOT_DIR := pkg/apis
GPU_API_GO_PKG := github.com/ihcsim/k8s-dra/pkg/apis
GPU_OPENAPI_GO_PKG := github.com/ihcsim/k8s-dra/pkg/openapi

BOILERPLATE_FILE := hack/boilerplate.go.txt

build: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build ./...

test: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test ./...

tidy:
	go mod tidy

codegen:
	source ./hack/kube-codegen.sh && \
	kube::codegen::gen_helpers --boilerplate $(BOILERPLATE_FILE) $(API_ROOT_DIR) && \
	kube::codegen::gen_openapi  \
		--boilerplate $(BOILERPLATE_FILE) \
		--output-dir $(API_ROOT_DIR) \
		--output-pkg $(GPU_OPENAPI_GO_PKG) \
		--report-filename $(API_ROOT_DIR)/openapi-report.txt \
		--update-report \
		$(API_ROOT_DIR) && \
	kube::codegen::gen_client \
		--boilerplate $(BOILERPLATE_FILE) \
		--output-dir $(API_ROOT_DIR) \
		--output-pkg $(GPU_API_GO_PKG) \
		--with-applyconfig \
		--with-watch \
		$(API_ROOT_DIR)
