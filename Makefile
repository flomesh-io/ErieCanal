# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

include VERSION
export $(shell sed 's/=.*//' VERSION)

# APP_VERSION represents the manager, proxy-init versions.
# This value must be updated to the release tag of the most recent release, a change that must
# occur in the release commit.
export PROJECT_NAME = erie-canal
# Build-time variables to inject into binaries
export SIMPLE_VERSION = $(shell (test "$(shell git describe --tags)" = "$(shell git describe --abbrev=0 --tags)" && echo $(shell git describe --tags)) || echo $(shell git describe --abbrev=0 --tags)+git)
export GIT_VERSION = $(shell git describe --dirty --tags --always)
export GIT_COMMIT = $(shell git rev-parse HEAD)
export BUILD_DATE ?= $(shell date +%Y-%m-%d-%H:%M-%Z)

# Build settings
export TOOLS_DIR = bin
#export SCRIPTS_DIR = tools/scripts
REPO = $(shell go list -m)
BUILD_DIR = bin

GO_ASMFLAGS ?= "all=-trimpath=$(shell dirname $(PWD))"
GO_GCFLAGS ?= "all=-trimpath=$(shell dirname $(PWD))"

LDFLAGS_COMMON =  \
	-X $(REPO)/pkg/version.Version=$(SIMPLE_VERSION) \
	-X $(REPO)/pkg/version.GitVersion=$(GIT_VERSION) \
	-X $(REPO)/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X $(REPO)/pkg/version.KubernetesVersion=v$(K8S_VERSION) \
	-X $(REPO)/pkg/version.ImageVersion=$(APP_VERSION) \
	-X $(REPO)/pkg/version.BuildDate=$(BUILD_DATE)

GO_LDFLAGS ?= "$(LDFLAGS_COMMON) -s -w"
#GO_BUILD_ARGS = -gcflags $(GO_GCFLAGS) -asmflags $(GO_ASMFLAGS) -ldflags $(GO_LDFLAGS)
GO_BUILD_ARGS = -ldflags $(GO_LDFLAGS)

export GO111MODULE = on
export CGO_ENABLED = 0
#export GOPROXY=https://goproxy.io
export PATH := $(PWD)/$(BUILD_DIR):$(PWD)/$(TOOLS_DIR):$(PATH)

export BUILD_IMAGE_REPO = flomesh
export IMAGE_TARGET_LIST = manager proxy-init ingress-pipy
IMAGE_PLATFORM = linux/amd64
ifeq ($(shell uname -m),arm64)
	IMAGE_PLATFORM = linux/arm64
endif

export CHART_COMPONENTS_DIR = charts/$(PROJECT_NAME)/components
export SCRIPTS_TAR = $(CHART_COMPONENTS_DIR)/scripts.tar.gz

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
#CRD_OPTIONS ?= "crd:trivialVersions=false,preserveUnknownFields=false"
CRD_OPTIONS ?= "crd:generateEmbeddedObjectMeta=true"
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.

##@ Development

.PHONY: manifests
manifests: controller-gen kustomize ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=charts/$(PROJECT_NAME)/apis
	$(KUSTOMIZE) build charts/erie-canal/apis/ -o charts/erie-canal/apis/flomesh.io_mcs-api.yaml
	rm -fv charts/erie-canal/apis/flomesh.io_serviceexports.yaml \
		charts/erie-canal/apis/flomesh.io_serviceimports.yaml \
		charts/erie-canal/apis/flomesh.io_globaltrafficpolicies.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

.PHONY: go-mod-tidy
go-mod-tidy:
	./hack/go-mod-tidy.sh

.PHONY: verify-codegen
verify-codegen:
	./hack/verify-codegen.sh

.PHONY: check-scripts
check-scripts:
	./hack/check-scripts.sh

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager, proxy-init, ingress-pipy with release args, the result will be optimized.
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_ARGS) -o $(BUILD_DIR)/erie-canal ./cli
	go build $(GO_BUILD_ARGS) -o $(BUILD_DIR) ./cmd/{manager,proxy-init,ingress-pipy}

.PHONY: build/manager build/proxy-init build/ingress-pipy
build/manager build/proxy-init build/ingress-pipy:
	go build $(GO_BUILD_ARGS) -o $(BUILD_DIR)/$(@F) ./cmd/$(@F)

##@ Development

.PHONY: codegen
codegen: ## Generate ClientSet, Informer, Lister and Deepcopy code for Flomesh CRD
	./hack/update-codegen.sh

.PHONY: package-scripts
package-scripts: ## Tar all repo initializing scripts
	tar -C $(CHART_COMPONENTS_DIR)/ -zcvf $(SCRIPTS_TAR) scripts/

.PHONY: charts-tgz-rel
charts-tgz-rel: helm
	export PACKAGED_APP_VERSION=$(APP_VERSION) HELM_BIN=$(LOCALBIN)/helm && ./hack/gen-charts-tgz.sh

.PHONY: charts-tgz-dev
charts-tgz-dev: helm
	export PACKAGED_APP_VERSION=$(APP_VERSION)-dev HELM_BIN=$(LOCALBIN)/helm && ./hack/gen-charts-tgz.sh

.PHONY: dev
dev: charts-tgz-dev manifests build kustomize ## Create dev commit changes to commit & Write dev commit changes.

.PHONY: build_docker_setup
build_docker_setup:
	docker run --rm --privileged tonistiigi/binfmt:latest --install all
	docker buildx create --name $(PROJECT_NAME)
	docker buildx use $(PROJECT_NAME)
	docker buildx inspect --bootstrap

.PHONY: rm_docker_builder
rm_docker_builder:
	docker buildx rm $(PROJECT_NAME) || true

.PHONY: build_docker_push
build_docker_push: build_docker_setup build_push_images rm_docker_builder

.PHONY: build_push_images
build_push_images: $(foreach i,$(IMAGE_TARGET_LIST),build_push_image/$(i))

build_push_image/%:
	docker buildx build --platform $(IMAGE_PLATFORM) \
		-t $(BUILD_IMAGE_REPO)/$(PROJECT_NAME)-$*:$(APP_VERSION)-dev \
		-f ./dockerfiles/$*/Dockerfile \
		--build-arg DISTROLESS_TAG=debug \
		--push \
		--cache-from "type=local,src=.buildcache" \
		--cache-to "type=local,dest=.buildcache" \
		.

.PHONY: go-lint
go-lint:
	docker run --rm -v $$(pwd):/app -w /app golangci/golangci-lint:v1.45.2-alpine golangci-lint run --config .golangci.yml


##@ Release

.PHONY: check_release_version
check_release_version:
ifeq (,$(RELEASE_VERSION))
	$(error "RELEASE_VERSION must be set to a release tag")
endif
ifneq ("$(RELEASE_VERSION)","v$(APP_VERSION)")
	$(error "APP_VERSION "v$(APP_VERSION)" must be updated to match RELEASE_VERSION "$(RELEASE_VERSION)" prior to creating a release commit")
endif

.PHONY: gh-release
gh-release: charts-tgz-rel ## Using goreleaser to Release target on Github.
ifeq (,$(GIT_VERSION))
	$(error "GIT_VERSION must be set to a git tag")
endif
	go install github.com/goreleaser/goreleaser@v1.13.0
	GORELEASER_CURRENT_TAG=$(GIT_VERSION) goreleaser release --rm-dist --parallelism 5

.PHONY: gh-release-snapshot
gh-release-snapshot: charts-tgz-rel
ifeq (,$(GIT_VERSION))
	$(error "GIT_VERSION must be set to a git tag")
endif
	GORELEASER_CURRENT_TAG=$(GIT_VERSION) goreleaser release --snapshot --rm-dist --parallelism 5 --debug

.PHONY: gh-build-snapshot
gh-build-snapshot: charts-tgz-rel
ifeq (,$(GIT_VERSION))
	$(error "GIT_VERSION must be set to a git tag")
endif
	GORELEASER_CURRENT_TAG=$(GIT_VERSION) goreleaser build --snapshot --rm-dist --parallelism 5 --debug


.PHONY: pre-release
pre-release: check_release_version manifests generate fmt vet kustomize  ## Create release commit changes to commit & Write release commit changes.


.PHONY: release
VERSION_REGEXP := ^v[0-9]+\.[0-9]+\.[0-9]+(\-(alpha|beta|rc)\.[0-9]+)?$
release: ## Create a release tag, push to git repository and trigger the release workflow.
ifeq (,$(RELEASE_VERSION))
	$(error "RELEASE_VERSION must be set to tag HEAD")
endif
ifeq (,$(shell [[ "$(RELEASE_VERSION)" =~ $(VERSION_REGEXP) ]] && echo 1))
	$(error "Version $(RELEASE_VERSION) must match regexp $(VERSION_REGEXP)")
endif
	git tag --sign --message "$(PROJECT_NAME) $(RELEASE_VERSION)" $(RELEASE_VERSION)
	git verify-tag --verbose $(RELEASE_VERSION)
	git push origin --tags

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
HELM ?= $(LOCALBIN)/helm
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.6
HELM_VERSION ?= v3.11.1
CONTROLLER_TOOLS_VERSION ?= v0.11.3

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	[ -f $(KUSTOMIZE) ] || curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

HELM_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3"
.PHONY: helm
helm: $(HELM) ## Download kustomize locally if necessary.
$(HELM): $(LOCALBIN)
	[ -f $(HELM) ] || curl -s $(HELM_INSTALL_SCRIPT) | HELM_INSTALL_DIR=$(LOCALBIN) bash -s -- --version $(HELM_VERSION) --no-sudo

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	[ -f $(CONTROLLER_GEN) ] || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	[ -f $(ENVTEST) ] || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.DEFAULT_GOAL := help
.PHONY: help
help: ## Show this help screen.
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
