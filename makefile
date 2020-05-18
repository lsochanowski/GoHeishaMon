  GOCMD=go
  GOBUILD=$(GOCMD) build
  GOCLEAN=$(GOCMD) clean
  GOTEST=$(GOCMD) test
  GOGET=$(GOCMD) get
  BINARY_NAME=GoHeishaMon
  BINARY_UNIX=$(BINARY_NAME)_AMD64
  BINARY_MIPS=$(BINARY_NAME)_MIPS
  BINARY_ARM=$(BINARY_NAME)_ARM
  BINARY_MIPSUPX=$(BINARY_NAME)_MIPSUPX





      all: test build-linux build-mips build-rpi
    build: 
	$(GOBUILD) -o $(BINARY_NAME) -v
    test: 
	$(GOTEST) -v ./...
    clean: 
	$(GOCLEAN)
	rm -f dist/*
    run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
    deps:
	$(GOGET) github.com/eclipse/paho.mqtt.golang
	$(GOGET) go.bug.st/serial
	$(GOGET) github.com/BurntSushi/toml
	$(GOGET) github.com/rs/xid
	mkdir dist
    build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_UNIX)
    build-mips:
	CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat $(GOBUILD) -ldflags "-s -w" -a -o dist/$(BINARY_MIPS)
    build-rpi:
	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o dist/$(BINARY_ARM)
    upx:
	upx -f --brute -o dist/$(BINARY_MIPSUPX) dist/$(BINARY_MIPS)
    install:
	scp dist/GoHeishaMon_MIPSUPX root@172.17.136.82:/usr/bin/
	ssh root@172.17.136.82 reboot
    compilesquash:
	cp dist/$(BINARY_MIPSUPX) OS/RootFS/usr/bin/$(BINARY_MIPSUPX)
	mksquashfs OS/RootFS dist/openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin -comp xz -noappend -always-use-fragments

