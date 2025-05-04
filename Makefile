.DEFAULT_GOAL := build

VERSION 	?= $(shell git describe --tags --always --dirty)
BUILD_FLAGS ?= -v
ARCH        ?= $(shell go env GOARCH)
GOARCH      ?= $(ARCH)
OS          ?= $(shell go env GOOS)
GOOS        ?= $(OS)
PACKAGE_DIR ?= packages

.PHONY: check
check: test lint

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: generate
generate:
	GOARCH=$(ARCH) GOOS=$(OS) go generate -v ./...

.PHONY: build
build: bin/zfs-rclone

.PHONY: clean
clean:
	rm -rf ./bin
	rm -rf $(PACKAGE_DIR)

bin/zfs-rclone: generate
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o bin/zfs-rclone $(BUILD_FLAGS) ./main.go

.PHONY: package
package:
	$(MAKE) build GOOS=linux
	mkdir -p $(PACKAGE_DIR)
	VERSION=$(VERSION) nfpm package \
		--config nfpm.yaml \
		--target $(PACKAGE_DIR)/zfs-rclone_$(VERSION)_$(ARCH).deb

.PHONY: package-all
package-all: package-amd64 package-arm64

.PHONY: package-amd64
package-amd64:
	$(MAKE) package ARCH=amd64 GOARCH=amd64

.PHONY: package-arm64
package-arm64:
	$(MAKE) package ARCH=arm64 GOARCH=arm64