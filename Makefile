# update VERSION before running `make release`
VERSION := 0.0.0

BASE_DIR := $(shell pwd)
BUILD_DIR := $(BASE_DIR)/build
DIST_DIR := $(BASE_DIR)/dist

export GOOS ?= linux
export GOARCH ?= amd64
export GOPATH ?= $(BASE_DIR)/go

BUILD_FLAGS := -s -w
ifeq ($(GOOS), windows)
	BUILD_FLAGS += -H=windowsgui
	BIN_EXT := .exe
endif

ifeq ($(GOOS), darwin)
	OS := macos
else
	OS := $(GOOS)
endif

BIN_NAME := buster-client-v$(VERSION)-$(OS)-$(GOARCH)$(BIN_EXT)

GOBINDATA := $(GOPATH)/bin/go-bindata

.PHONY: build
build:
	@echo "Building package (version: $(VERSION), os: $(GOOS), arch: $(GOARCH))"
	@cd cmd/client || exit 1 && go build -ldflags "-s -w -X=main.buildVersion=$(VERSION)" -o $(BUILD_DIR)/bin/buster$(BIN_EXT)

.PHONY: installer
installer: build $(GOBINDATA)
	@echo "Building installer (version: $(VERSION), os: $(GOOS), arch: $(GOARCH))"
	@rm -rf $(BUILD_DIR)/src && mkdir -p $(BUILD_DIR)/src && cp -r cmd/installer $(BUILD_DIR)/src
	@cd $(BUILD_DIR)/bin || exit 1 && $(GOBINDATA) -mode 0755 -o $(BUILD_DIR)/src/installer/appbin.go buster$(BIN_EXT)
	@cd $(BUILD_DIR)/src/installer || exit 1 && go build -ldflags "$(BUILD_FLAGS)" -o $(DIST_DIR)/$(BIN_NAME)

.PHONY: release
release:
	@echo "Releasing version ($(VERSION))"
	@git add Makefile && git commit -m "chore(release): $(VERSION)" && git tag v$(VERSION)
	@git push --follow-tags origin master

$(GOBINDATA):
	@GO111MODULE=off go get -u github.com/dessant/go-bindata/...
