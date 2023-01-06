GIT_VERSION?=$(shell git describe --tags --dirty --abbrev=0 2>/dev/null || git symbolic-ref --short HEAD)
GOPACKAGE=$(shell go list -m)

ldflags+=-w -s
ldflags+=-X '${GOPACKAGE}.gitVersion=${GIT_VERSION}'

define build
	@echo "Building ${1}/${2}"
	@CGO_ENABLED=0 GOOS=${1} GOARCH=$(2) go build -gcflags=all="-N -l" -ldflags="${ldflags}" -o bin/alidnsctl-$(1)-$(2) ${GOPACKAGE}/cmd/alidnsctl
endef

all: build-all

OS:=$(shell go env GOOS)
ARCH:=$(shell go env GOARCH)
build: ## Build binaries.
	$(call build,${OS},${ARCH})

build-all:
	$(call build,linux,amd64)
	$(call build,linux,arm64)
	$(call build,darwin,amd64)
	$(call build,darwin,arm64)
	$(call build,windows,amd64)