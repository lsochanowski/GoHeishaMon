GOCMD=go
GOBUILD=$(GOCMD) build
.POSIX:
.PHONY: *

GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=GoHeishaMon
BINARY_UNIX=$(BINARY_NAME)_AMD64
BINARY_MIPS=$(BINARY_NAME)_MIPS
BINARY_ARM=$(BINARY_NAME)_ARM
BINARY_MIPSUPX=$(BINARY_NAME)_MIPSUPX

.DEFAULT: help
help:	## show this help menu.
	@echo "Usage: make [TARGET ...]"
	@echo ""
	@@grep -hE "#[#]" $(MAKEFILE_LIST) | sed -e 's/\\$$//' | awk 'BEGIN {FS = "[:=].*?#[#] "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

all:    ## test and build for all platforms
all: test build-linux build-mips build-rpi

build:  ## build binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:   ## run go tests
test:
	$(GOTEST) -v ./...

clean:  ## clean dist dir
clean:
	$(GOCLEAN)
	rm -f dist/*

run:    ## run main program
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:   ## create dist
deps:
	$(GOGET) github.com/eclipse/paho.mqtt.golang
	$(GOGET) go.bug.st/serial
	$(GOGET) github.com/BurntSushi/toml
	$(GOGET) github.com/rs/xid
	mkdir dist

build-linux:    ## build for linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_UNIX)

build-mips: ## build for MIPS
build-mips:
	CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat $(GOBUILD) -ldflags "-s -w" -a -o dist/$(BINARY_MIPS)

build-rpi:  ## build for ARM
build-rpi:
	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o dist/$(BINARY_ARM)

upx:    ## package binary
upx:
	upx -f --brute -o dist/$(BINARY_MIPSUPX) dist/$(BINARY_MIPS)

install:    ## install in TARGET_HOST
install:
	scp dist/GoHeishaMon_MIPSUPX root@${TARGET_HOST}:/usr/bin/
	ssh root@${TARGET_HOST} reboot

compilesquash: ## create root file system
compilesquash:
	cp dist/$(BINARY_MIPSUPX) OS/RootFS/usr/bin/$(BINARY_MIPSUPX)
	mksquashfs OS/RootFS dist/openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin -comp xz -noappend -always-use-fragments

