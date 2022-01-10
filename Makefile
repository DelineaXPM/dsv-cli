BUILD = $(shell date +%Y%m%d%H%M)
VERSION = $(shell git describe --always --dirty)

PKGNAME = dsv
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

LDFLAGS = -X thy/version.Version=$(VERSION) -X thy/version.Build=$(BUILD)
LDFLAGS_REL := $(LDFLAGS) -s -w

clean:
	$(shell rm -rf bin)

test:
	go test ./...

build:
	GO111MODULE=on go build -ldflags="$(LDFLAGS)" -o $(PKGNAME)$(EXE_SUFFIX)

build-test:
	GO111MODULE=on go test -c -covermode=count -coverpkg ./... -o $(PKGNAME)$(EXE_SUFFIX).test

build-release:
	GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o $(PKGNAME)$(EXE_SUFFIX)

build-release-all:
	# Note: need to set CGO_ENABLED = 0 for linux build to work on barebones docker containers (like scratch which doesnt have c std libs)
	CGO_ENABLED=0  GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o bin/$(VERSION)/$(PKGNAME)-win-x64.exe
	CGO_ENABLED=0  GOOS=windows GOARCH=386   GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o bin/$(VERSION)/$(PKGNAME)-win-x86.exe
	CGO_ENABLED=0  GOOS=linux GOARCH=amd64   GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o bin/$(VERSION)/$(PKGNAME)-linux-x64
	CGO_ENABLED=0  GOOS=linux GOARCH=386     GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o bin/$(VERSION)/$(PKGNAME)-linux-x86
	CGO_ENABLED=0  GOOS=darwin GOARCH=amd64  GO111MODULE=on go build -ldflags="$(LDFLAGS_REL)" -o bin/$(VERSION)/$(PKGNAME)-darwin-x64

create-checksum:
	$(shell cd bin/$(VERSION); for file in *; do sha256sum $$file > $$file-sha256.txt; done)

TEMPLATE = '{"latest":"$(VERSION)","links":\
{"darwin/amd64":"https://dsv.thycotic.com/downloads/cli/$(VERSION)/$(PKGNAME)-darwin-x64",\
"linux/amd64":"https://dsv.thycotic.com/downloads/cli/$(VERSION)/$(PKGNAME)-linux-x64",\
"linux/386":"https://dsv.thycotic.com/downloads/cli/$(VERSION)/$(PKGNAME)-linux-x86",\
"windows/amd64":"https://dsv.thycotic.com/downloads/cli/$(VERSION)/$(PKGNAME)-win-x64.exe",\
"windows/386":"https://dsv.thycotic.com/downloads/cli/$(VERSION)/$(PKGNAME)-win-x86.exe"}}'

capture-latest-version:
	$(shell echo $(TEMPLATE) > bin/cli-version.json)

.DEFAULT_GOAL := build
