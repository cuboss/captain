# captain-server version
VERSION = v0.1.0

OUTPUT_DIR=bin
LDFLAGS=$(kube::version::ldflags)
GOBINARY=go
CAPTAIN_APISERVER_BUILDPATH=./cmd/captain-server

IMAGE_NAME=cuboss/captain-server

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all

all: test captain-server

.PHONY: binary image

binary: | captain-server ; $(info $(M)...Build all f binary.) @ ## Build all of binary

# build captain-server binary
captain-server: ; $(info $(M)...Begin to build captain-apiserver binary.)  @ ## Build captain-apiserver.
	GOOS=${BUILD_GOOS} CGO_ENABLED=0 GOARCH=${BUILD_GOARCH} ${GOBINARY} build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/captain-server" ${CAPTAIN_APISERVER_BUILDPATH}

image: binary
	docker build -t ${IMAGE_NAME}:${VERSION} .

