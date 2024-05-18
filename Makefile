SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

API_ROOT_DIR := pkg/apis
API_GO_PKG := github.com/ihcsim/k8s-dra/pkg/apis
OPENAPI_GO_PKG := github.com/ihcsim/k8s-dra/pkg/openapi

BOILERPLATE_FILE := hack/boilerplate.go.txt

build: tidy lint
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build ./...

test: tidy lint
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

codegen:
	rm -rf pkg/apis/applyconfiguration pkg/apis/clientset pkg/apis/informers pkg/apis/listers && \
	source ./hack/kube-codegen.sh && \
	kube::codegen::gen_helpers --boilerplate $(BOILERPLATE_FILE) $(API_ROOT_DIR) && \
	kube::codegen::gen_openapi --boilerplate $(BOILERPLATE_FILE) \
		--output-dir $(API_ROOT_DIR)/openapi \
		--output-pkg $(OPENAPI_GO_PKG) \
		--report-filename $(API_ROOT_DIR)/openapi-report.txt \
		--update-report \
		$(API_ROOT_DIR) && \
	kube::codegen::gen_client --boilerplate $(BOILERPLATE_FILE) \
		--output-dir $(API_ROOT_DIR) \
		--output-pkg $(API_GO_PKG) \
		--with-applyconfig \
		--with-watch \
		--plural-exceptions "GPUClassParameters:GPUClassParameters,GPUClaimParameters:GPUClaimParameters" \
		$(API_ROOT_DIR)

codegen-verify:
	@srcdir=$$(pwd) && \
	tmpdir=$$(mktemp -d -t k8s-dra.XXXXXX ) && \
	cp -a . $${tmpdir} && \
	pushd $${tmpdir} && \
	$(MAKE) -s codegen && \
	diff -Naupr "$${srcdir}" "$${tmpdir}" || ret="$$?" && \
	rm -rf $${tmpdir} && \
	echo "Removed temporary diff folder at $${tmpdir}." && \
	[[ $${ret} -eq 0 ]] || { echo "CRD APIs is outdated. Please run 'make codegen'."; exit 1; }
