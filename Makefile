PKGNAME := dsv
ARTIFACT_DIRECTORY := .artifacts/builds/
BIN_DIRECTORY := .artifacts/builds/bin/
TEST_DIRECTORY :=.artifacts/test/
ifneq ($(CONSTANTS_CLINAME),)
	PKGNAME = $(CONSTANTS_CLINAME)
endif

ifeq ($(OS),Windows_NT)
	EXE_SUFFIX = .exe
else
	ifeq ($(shell uname), Linux)
		EXE_SUFFIX =
	endif
endif

VERSION     = $(shell git describe --always --dirty --tags)
BUILD_DATE  = $(shell date +%s)
GIT_COMMIT  = $(shell git rev-parse HEAD)

LDFLAGS     = -X thy/version.Version=$(VERSION)
LDFLAGS     += -X thy/version.BuildDate=$(BUILD_DATE)
LDFLAGS     += -X thy/version.GitCommit=$(GIT_COMMIT)
LDFLAGS_REL = $(LDFLAGS) -s -w

.DEFAULT_GOAL := build

clean:
	@echo '❗ [DEPRECATED] REPlACED BY mage clean'
	rm -rf "$(BIN_DIRECTORY)"
	rm -rf "$(TEST_DIRECTORY)"
	mkdir -p "$(BIN_DIRECTORY)" || echo "✔️ $(BIN_DIRECTORY) already exists"
	mkdir -p "$(TEST_DIRECTORY)" || echo "✔️ $(TEST_DIRECTORY) already exists"

test:
	@echo '❗ [DEPRECATED] REPlACED BY "mage go:testsum ./..."'
	go test -v ./...

e2e-test:
	@echo '❗ [DEPRECATED] REPlACED BY "GOTEST_FLAGS='-tags=endtoend' mage go:testsum ./tests/e2e/..."'
	go clean -testcache
	go test -v -tags=endtoend ./tests/e2e

build:
	@echo '❗ [DEPRECATED] REPlACED BY mage build'
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags="$(LDFLAGS)" -o $(PKGNAME)$(EXE_SUFFIX)

build-test:
	@echo '❗ [DEPRECATED] REPlACED BY mage build go:test'
	CGO_ENABLED=0 GO111MODULE=on go test -c -covermode=count -coverpkg ./... -o $(PKGNAME)$(EXE_SUFFIX).test

build-release:
	@echo '❗ [DEPRECATED] REPlACED BY mage release'
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o $(BIN_DIRECTORY)$(PKGNAME)$(EXE_SUFFIX)

build-release-all: status
	@echo '❗ [DEPRECATED] REPlACED BY mage release'
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-win-x64.exe"
	CGO_ENABLED=0 GOOS=windows GOARCH=386   GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-win-x86.exe"
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-linux-x64"
	CGO_ENABLED=0 GOOS=linux   GOARCH=386   GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-linux-x86"
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-darwin-x64"
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o "$(BIN_DIRECTORY)$(VERSION)/$(PKGNAME)-darwin-arm64"

create-checksum:
	@echo '❗ [DEPRECATED] REPlACED BY mage release, pending s3 asset upload'
	$(shell cd $(BIN_DIRECTORY)$(VERSION); for file in *; do sha256sum $$file > $$file-sha256.txt; done)

TEMPLATE = '{"latest":"$(VERSION)","links":\
{"darwin/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-darwin-x64",\
"darwin/arm64":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-darwin-arm64",\
"linux/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-linux-x64",\
"linux/386":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-linux-x86",\
"windows/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-win-x64.exe",\
"windows/386":"https://dsv.secretsvaultcloud.com/downloads/cli/$(VERSION)/$(PKGNAME)-win-x86.exe"}}'

capture-latest-version:
	echo $(TEMPLATE) > $(BIN_DIRECTORY)cli-version.json
status:
	echo "---------- GIT STATUS ----------"
	git status
	echo "--------------------------------"
.PHONY: clean test e2e-test \
		build build-test build-release build-release-all \
		create-checksum capture-latest-version status
