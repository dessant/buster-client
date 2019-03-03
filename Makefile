VERSION := $(shell sed -n 's/.*"version": "\(.*\)",/\1/p' package.json)

BASE_DIR := $(shell pwd)
BUILD_DIR := $(BASE_DIR)/build
DIST_DIR := $(BASE_DIR)/dist

export GOOS ?= linux
export GOARCH ?= amd64
export GOPATH ?= $(BASE_DIR)/go

SETUP_LDFLAGS := -s -w
ifeq ($(GOOS), windows)
	SETUP_LDFLAGS += -H=windowsgui
	BIN_EXT := .exe
endif

ifeq ($(GOOS), darwin)
	OS := macos
else
	OS := $(GOOS)
endif

BIN_SUFFIX := v$(VERSION)-$(OS)-$(GOARCH)$(BIN_EXT)

GOBINDATA := $(GOPATH)/bin/go-bindata
STANDARD_VERSION := ~/.yarn/bin/standard-version

.PHONY: build
build:
	@echo "Building client (version: $(VERSION), os: $(GOOS), arch: $(GOARCH))"
	@cd cmd/client || exit 1 && go build -ldflags "-s -w -X=main.buildVersion=$(VERSION)" -o $(DIST_DIR)/buster-client-$(BIN_SUFFIX)

.PHONY: setup
setup: build $(GOBINDATA)
	@echo "Building setup (version: $(VERSION), os: $(GOOS), arch: $(GOARCH))"
	@rm -rf $(BUILD_DIR)/src && mkdir -p $(BUILD_DIR)/src && cp -r cmd lib $(BUILD_DIR)/src
	@mkdir -p $(BUILD_DIR)/bin && cp $(DIST_DIR)/buster-client-$(BIN_SUFFIX) $(BUILD_DIR)/bin/buster-client$(BIN_EXT)
	@cd $(BUILD_DIR)/bin || exit 1 && $(GOBINDATA) -mode 0755 -o $(BUILD_DIR)/src/cmd/setup/appbin.go buster-client$(BIN_EXT)
	@cd $(BUILD_DIR)/src/cmd/setup || exit 1 && go build -ldflags "$(SETUP_LDFLAGS)" -o $(DIST_DIR)/buster-client-setup-$(BIN_SUFFIX)

.PHONY: release
release: $(STANDARD_VERSION)
	@$(STANDARD_VERSION)
	@git push --follow-tags origin master

$(GOBINDATA):
	@GO111MODULE=off go get -u github.com/dessant/go-bindata/...

$(STANDARD_VERSION):
	@yarn global add standard-version
